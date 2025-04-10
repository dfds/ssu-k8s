package controller

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/git"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const capabilityDataNamespace = "capability-data"

func ReconcileCapabilityResources(ctx context.Context, client client.Client, capability model.Capability, ns string, repo *git.Repo) error {
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	err = repo.Add(model.Capability{
		Name: capability.Name,
		Id:   capability.Id,
	}, conf.Kubernetes.ClusterName)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func ReconcileCapabilityDeploymentToken(ctx context.Context, client client.Client, capability model.Capability, ns string) error {
	fmt.Println(capability)
	return nil
}
