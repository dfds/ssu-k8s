package controller

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/git"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/operator/misc"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SecretReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Repo   *git.Repo
}

// Reconcile TODO: Add reconcile logic for Capability namespaces
// Reconcile Capability namespaces must be labelled with the key "dfds.cloud/capability"
// Reconcile You can opt out of reconciling a namespace resource by setting a label of "dfds.cloud/reconcile: false"
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch secret obj
	secretObj := &v1.Secret{}
	err := r.Client.Get(ctx, req.NamespacedName, secretObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logging.Logger.Debug("secret resource not found. Ignoring since object must be deleted", zap.Error(err))
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logging.Logger.Debug("Failed to get secret", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Not a Capability secret, skip
	if _, ok := secretObj.Labels[misc.LabelCapabilityKey]; !ok {
		return ctrl.Result{}, nil
	}

	// Reconcile disabled, skip
	if _, ok := secretObj.Labels[misc.LabelReconcileKey]; ok {
		if secretObj.Labels[misc.LabelReconcileKey] == "false" {
			return ctrl.Result{}, nil
		}
	}

	// Secret is not of type "deployment-token", skip
	if _, ok := secretObj.Labels[misc.LabelTypeKey]; ok {
		if secretObj.Labels[misc.LabelTypeKey] != "deployment-token" {
			return ctrl.Result{}, nil
		}
	}

	// Check that Secret has enabled ssm-secrets feature
	featuresOptIns := misc.GetFeaturesFromAnnotation(secretObj.Annotations)
	if _, ok := featuresOptIns["ssm-secrets"]; !ok {
		return ctrl.Result{}, nil
	}

	// Fetch ns obj
	nsObj := &v1.Namespace{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: "",
		Name:      secretObj.Labels[misc.LabelCapabilityKey],
	}, nsObj)
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

	logging.Logger.Debug("Reconciling Secret: " + secretObj.Name)
	fmt.Println("Reconciling Secret: " + secretObj.Name)

	//err = ReconcileCapabilityResources(ctx, r.Client, model.Capability{
	//	Name: "",
	//	Id:   secretObj.Name,
	//}, secretObj.Name, r.Repo)
	//if err != nil {
	//	logging.Logger.Error("Failed to reconcile namespace child resources", zap.Error(err))
	//}

	err = ReconcileCapabilityDeploymentToken(ctx, r.Client, model.Capability{
		Name:         "",
		Id:           secretObj.Name,
		ContextId:    nsObj.Labels[misc.LabelContextIdKey],
		AwsAccountId: nsObj.Labels[misc.LabelAwsAccountKey],
	}, secretObj.Name)
	if err != nil {
		logging.Logger.Error("Failed to reconcile secret", zap.Error(err))
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretPredicate := predicate.Funcs{
		CreateFunc: func(e event.TypedCreateEvent[client.Object]) bool {
			return hasLabel(e.Object, "dfds.cloud/capability")
		},
		DeleteFunc: func(e event.TypedDeleteEvent[client.Object]) bool {
			return hasLabel(e.Object, "dfds.cloud/capability")
		},
		UpdateFunc: func(e event.TypedUpdateEvent[client.Object]) bool {
			return hasLabel(e.ObjectNew, "dfds.cloud/capability")
		},
		GenericFunc: func(e event.TypedGenericEvent[client.Object]) bool {
			return hasLabel(e.Object, "dfds.cloud/capability")
		},
	}

	return ctrl.NewControllerManagedBy(mgr).WithOptions(controller.Options{
		MaxConcurrentReconciles: 100,
	}).For(&v1.Secret{}).WithEventFilter(secretPredicate).Complete(r)
}

func hasLabel(obj client.Object, labelKey string) bool {
	labels := obj.GetLabels()
	_, exists := labels[labelKey]
	return exists
}
