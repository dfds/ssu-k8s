package controller

import (
	"context"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const capabilityDataNamespace = "capability-data"

func ReconcileCapabilityResources(ctx context.Context, client client.Client, capability model.Capability, ns string) error {
	//_, err := actions.GetObject[*v1.ServiceAccount](ctx, client, types.NamespacedName{
	//	Namespace: capabilityDataNamespace,
	//	Name:      capability.Id,
	//}, &v1.ServiceAccount{})
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		// serviceAccount not found, but is supposed to exist, creating
	//		logging.Logger.Debug(fmt.Sprintf("Capability %s missing serviceAccount, creating", capability.Id))
	//		err = actions.CreateServiceAccount(ctx, client, capability, ns)
	//		if err != nil {
	//			return err
	//		}
	//	} else {
	//		// Error reading the object - requeue the request.
	//		logging.Logger.Debug("Failed to get serviceAccount", zap.Error(err))
	//		return err
	//	}
	//}

	return nil
}
