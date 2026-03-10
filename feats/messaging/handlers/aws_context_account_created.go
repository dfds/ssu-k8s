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

type CapabilityMetadata struct {
	CostCentre          string `json:"dfds.cost.centre"`
	Env                 string `json:"dfds.env"`
	BusinessCapability  string `json:"dfds.other.businessCapability"`
	DataClassification  string `json:"dfds.data.classification"`
	ServiceCriticality  string `json:"dfds.service.criticality"`
	ServiceAvailability string `json:"dfds.service.availability"`
}

type AWSContextAccountCreated struct {
	AccountId        string               `json:"accountId"`
	CapabilityId     string               `json:"capabilityId"`
	CapabilityName   string               `json:"capabilityName"`
	CapabilityRootId string               `json:"capabilityRootId"`
	ContextId        string               `json:"contextId"`
	ContextName      string               `json:"contextName"`
	RoleArn          string               `json:"roleArn"`
	RoleEmail        string               `json:"roleEmail"`
	Metadata         *CapabilityMetadata  `json:"metadata,omitempty"`
}

type CapabilityMetadataUpdated struct {
	CapabilityId     string              `json:"capabilityId"`
	CapabilityRootId string              `json:"capabilityRootId"`
	Metadata         CapabilityMetadata  `json:"metadata"`
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

			labels := map[string]string{
				"dfds.cloud/capability":              msg.Payload.CapabilityRootId,
				"dfds.cloud/reconcile":               "true",
				"dfds.cloud/context-id":              msg.Payload.ContextId,
				"dfds.cloud/aws-account":             msg.Payload.AccountId,
				"pod-security.kubernetes.io/enforce": "baseline",
			}

			// Add metadata labels if present
			if msg.Payload.Metadata != nil {
				addMetadataLabels(labels, *msg.Payload.Metadata)
			}

			_, err := client.CoreV1().Namespaces().Create(ctx, &v1Core.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name:   msg.Payload.CapabilityRootId,
					Labels: labels,
				},
			}, v1.CreateOptions{})
			if err != nil {
				return err
			}

			// publish msg
			payload := model.EnvelopeWithPayload[events.K8sNamespaceCreatedAndAwsArnConnected]{
				EventName:      "k8s_namespace_created_and_aws_arn_connected",
				Version:        "1",
				XCorrelationId: "",
				XSender:        "ssu-k8s",
				Payload: events.K8sNamespaceCreatedAndAwsArnConnected{
					CapabilityId:  msg.Payload.CapabilityId,
					ContextId:     msg.Payload.ContextId,
					NamespaceName: msg.Payload.CapabilityRootId,
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

func CapabilityMetadataUpdatedHandler(ctx context.Context, event model.HandlerContext) error {
	logging.Logger.Info("capability_metadata_updated received")

	msg, err := messagingModel.SerialiseToEnvelopeWithPayload[CapabilityMetadataUpdated](event.Msg)
	if err != nil {
		return err
	}

	logger := logging.Logger.With(zap.String("handler", "capability_metadata_updated"), zap.String("capability_id", msg.Payload.CapabilityId))

	client, err := getK8sClient()
	if err != nil {
		return err
	}

	// Get the existing namespace
	ns, err := client.CoreV1().Namespaces().Get(ctx, msg.Payload.CapabilityRootId, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Warn("Namespace not found, skipping metadata update")
			return nil
		}
		logger.Error("Failed to get namespace", zap.Error(err))
		return err
	}

	// Initialize labels map if nil
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}

	// Update labels with metadata
	addMetadataLabels(ns.Labels, msg.Payload.Metadata)

	// Update the namespace
	_, err = client.CoreV1().Namespaces().Update(ctx, ns, v1.UpdateOptions{})
	if err != nil {
		logger.Error("Failed to update namespace labels", zap.Error(err))
		return err
	}

	logger.Info("Successfully updated namespace labels")
	return nil
}

func addMetadataLabels(labels map[string]string, metadata CapabilityMetadata) {
	if metadata.CostCentre != "" {
		labels["dfds.cost.centre"] = metadata.CostCentre
	}
	if metadata.Env != "" {
		labels["dfds.env"] = metadata.Env
	}
	if metadata.BusinessCapability != "" {
		labels["dfds.other.businessCapability"] = metadata.BusinessCapability
	}
	if metadata.DataClassification != "" {
		labels["dfds.data.classification"] = metadata.DataClassification
	}
	if metadata.ServiceCriticality != "" {
		labels["dfds.service.criticality"] = metadata.ServiceCriticality
	}
	if metadata.ServiceAvailability != "" {
		labels["dfds.service.availability"] = metadata.ServiceAvailability
	}
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
