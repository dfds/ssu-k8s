package handlers

import (
	"context"
	"go.dfds.cloud/messaging/kafka/model"
	"go.dfds.cloud/ssu-k8s/core/logging"
	messagingModel "go.dfds.cloud/ssu-k8s/feats/messaging/model"
	"go.uber.org/zap"
	v1Core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/env"
)

type AWSContextAccountCreated struct {
	AccountId        string `json:"accountId"`
	CapabilityId     string `json:"capabilityId"`
	CapabilityName   string `json:"capabilityName"`
	CapabilityRootId string `json:"capabilityRootId"`
	ContextId        string `json:"contextId"`
	ContextName      string `json:"contextName"`
	RoleArn          string `json:"roleArn"`
	RoleEmail        string `json:"roleEmail"`
}

func AwsContextAccountCreatedHandler(ctx context.Context, event model.HandlerContext) error {
	logging.Logger.Info("aws_context_account_created received")

	msg, err := messagingModel.SerialiseToEnvelopeWithPayload[AWSContextAccountCreated](event.Msg)
	if err != nil {
		return err
	}

	logger := logging.Logger.With(zap.String("handler", "aws_context_account_created"), zap.String("capability_id", msg.Payload.CapabilityId))

	client, err := getK8sClient()
	if err != nil {
		return err
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, msg.Payload.CapabilityRootId, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("Namespace missing, creating it")

			_, err := client.CoreV1().Namespaces().Create(ctx, &v1Core.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: msg.Payload.CapabilityRootId,
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
