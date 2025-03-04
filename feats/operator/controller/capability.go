package controller

import (
	"context"
	"errors"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/operator/actions"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const capabilityDataNamespace = "capability-data"

func ReconcileCapabilityResources(ctx context.Context, client client.Client, capability model.Capability, ns string) error {
	_, err := actions.GetObject[*v1.ServiceAccount](ctx, client, types.NamespacedName{
		Namespace: capabilityDataNamespace,
		Name:      capability.Id,
	}, &v1.ServiceAccount{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// serviceAccount not found, but is supposed to exist, creating
			logging.Logger.Debug(fmt.Sprintf("Capability %s missing serviceAccount, creating", capability.Id))
			err = actions.CreateServiceAccount(ctx, client, capability, ns)
			if err != nil {
				return err
			}
		} else {
			// Error reading the object - requeue the request.
			logging.Logger.Debug("Failed to get serviceAccount", zap.Error(err))
			return err
		}
	}

	return nil
}

func ReconcileFluxCapabilityResources(ctx context.Context, client client.Client, capability model.Capability, ns string) error {
	secret, err := actions.GetObject[*v1.Secret](ctx, client, types.NamespacedName{
		Namespace: capabilityDataNamespace,
		Name:      fmt.Sprintf("%s-token", capability.Id),
	}, &v1.Secret{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Secret not found, but is supposed to exist
			logging.Logger.Debug(fmt.Sprintf("Capability %s missing serviceAccount secret, Flux may be having issues or have not created a secret yet", capability.Id))
			return nil
		} else {
			// Error reading the object - requeue the request
			logging.Logger.Debug("Failed to get secret", zap.Error(err))
			return err
		}
	}

	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if !actions.IsCapabilityResource(secret.Labels) {
		return nil
	}

	// Retrieve k8s token
	if _, ok := secret.Data["token"]; !ok {
		logging.Logger.Error(fmt.Sprintf("Capability %s serviceAccount secret is missing token, unexpected", capability.Id))
		return errors.New("serviceAccount secret is missing token")
	}

	tokenString := string(secret.Data["token"])

	// Check if it already exists in AWS SSM

	// gen kubeconfig

	kubeconfig, err := model.GenerateKubeConfig(model.KubeConfigData{
		Name:      conf.Kubernetes.ClusterName,
		CaCert:    conf.Kubernetes.ClusterCaCert,
		Endpoint:  conf.Kubernetes.ClusterEndpoint,
		Token:     tokenString,
		Namespace: ns,
	})

	fmt.Println(string(kubeconfig))

	return nil
}
