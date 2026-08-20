package handlers

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.dfds.cloud/messaging/kafka/model"
	"go.dfds.cloud/ssu-k8s/core/logging"
	messagingModel "go.dfds.cloud/ssu-k8s/feats/messaging/model"
	"go.dfds.cloud/ssu-k8s/feats/messaging/model/events"
	"go.uber.org/zap"
	v1Core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/env"
)

type K8sNamespaceRequested struct {
	AccountId        string `json:"accountId"`
	CapabilityId     string `json:"capabilityId"`
	CapabilityRootId string `json:"capabilityRootId"`
	ContextId        string `json:"contextId"`
	NamespaceName    string `json:"namespaceName"`
}

func K8sNamespaceRequestedHandler(ctx context.Context, event model.HandlerContext) error {
	logging.Logger.Info("k8s_namespace_requested received")

	msg, err := messagingModel.SerialiseToEnvelopeWithPayload[K8sNamespaceRequested](event.Msg)
	if err != nil {
		return err
	}

	logger := logging.Logger.With(zap.String("handler", "k8s_namespace_requested"), zap.String("capability_id", msg.Payload.CapabilityId))

	const maxNamespaceLength = 63
	const maxNamespaceNameLength = 25
	const maxRootIdLength = maxNamespaceLength - maxNamespaceNameLength - 1 // -1 for hyphen
	const keepRootIdSuffix = 5

	namespacePart := msg.Payload.NamespaceName
	if len(namespacePart) > maxNamespaceNameLength {
		namespacePart = namespacePart[:maxNamespaceNameLength]
	}

	rootIdPart := msg.Payload.CapabilityRootId
	if len(rootIdPart) > maxRootIdLength {
		truncateAt := maxRootIdLength - keepRootIdSuffix
		rootIdPart = rootIdPart[:truncateAt] + rootIdPart[len(rootIdPart)-keepRootIdSuffix:]
	}

	namespaceName := rootIdPart + "-" + namespacePart

	client, err := getK8sClient()
	if err != nil {
		return err
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespaceName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Namespace missing, creating it")

			_, err := client.CoreV1().Namespaces().Create(ctx, &v1Core.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: namespaceName,
					Labels: map[string]string{
						"dfds.cloud/capability":              msg.Payload.CapabilityRootId,
						"dfds.cloud/reconcile":               "true",
						"dfds.cloud/context-id":              msg.Payload.ContextId,
						"dfds.cloud/aws-account":             msg.Payload.AccountId,
						"pod-security.kubernetes.io/enforce": "baseline",
					},
				},
			}, v1.CreateOptions{})
			if err != nil {
				return err
			}

			payload := model.EnvelopeWithPayload[events.K8sNamespaceCreatedAndAwsArnConnected]{
				EventName:      "k8s_namespace_created_and_aws_arn_connected",
				Version:        "1",
				XCorrelationId: "",
				XSender:        "ssu-k8s",
				Payload: events.K8sNamespaceCreatedAndAwsArnConnected{
					CapabilityId:  msg.Payload.CapabilityId,
					ContextId:     msg.Payload.ContextId,
					NamespaceName: namespaceName,
				},
			}

			serialised, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			err = event.Writer("build.selfservice.events.capabilities").WriteMessages(ctx, kafka.Message{Value: serialised})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if ns != nil {
		logger.Error("Received event for creating namespace that already exists")
	}

	return nil
}

func getK8sClient() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", env.GetString("KUBECONFIG", ""))
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
