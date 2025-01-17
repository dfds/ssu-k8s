package actions

import (
	"context"
	"go.dfds.cloud/ssu-k8s/feats/operator/misc"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetObject[T client.Object](ctx context.Context, client client.Client, nn types.NamespacedName, obj T) (T, error) {
	err := client.Get(ctx, nn, obj)
	if err != nil {
		return obj, err
	}

	return obj, nil
}

func AddCapabilityLabels(labels map[string]string, capability model.Capability) map[string]string {
	labels[misc.LabelCapabilityKey] = capability.Id

	return labels
}
