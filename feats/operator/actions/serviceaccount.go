package actions

import (
	"context"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	v1 "k8s.io/api/core/v1"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateServiceAccount(ctx context.Context, client client.Client, capability model.Capability, ns string) error {
	svcAcc := &v1.ServiceAccount{
		ObjectMeta: v2.ObjectMeta{
			Name:      capability.Id,
			Namespace: ns,
			Labels:    AddCapabilityLabels(make(map[string]string), capability),
		},
	}

	return client.Create(ctx, svcAcc)
}
