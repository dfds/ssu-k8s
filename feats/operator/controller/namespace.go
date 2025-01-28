package controller

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/operator/misc"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile TODO: Add reconcile logic for Capability namespaces
// Reconcile Capability namespaces must be labelled with the key "dfds.cloud/capability"
// Reconcile You can opt out of reconciling a namespace resource by setting a label of "dfds.cloud/reconcile: false"
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nsObj := &v1.Namespace{}

	err := r.Client.Get(ctx, req.NamespacedName, nsObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logging.Logger.Debug("namespace resource not found. Ignoring since object must be deleted", zap.Error(err))
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logging.Logger.Debug("Failed to get namespace", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Not a Capability namespace, skip
	if _, ok := nsObj.Labels[misc.LabelCapabilityKey]; !ok {
		return ctrl.Result{}, nil
	}

	// Reconcile disabled, skip
	if _, ok := nsObj.Labels[misc.LabelReconcileKey]; ok {
		if nsObj.Labels[misc.LabelReconcileKey] == "false" {
			return ctrl.Result{}, nil
		}
	}

	fmt.Println("Reconciling Namespace: " + nsObj.Name)
	//fmt.Println("labels:")
	//for k, v := range nsObj.Labels {
	//	fmt.Printf("  %s: %s\n", k, v)
	//}

	err = ReconcileFluxCapabilityResources(ctx, r.Client, model.Capability{
		Name: "",
		Id:   nsObj.Name,
	}, nsObj.Name)
	if err != nil {
		logging.Logger.Error("Failed to reconcile namespace child resources", zap.Error(err))
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Namespace{}).Complete(r)
}
