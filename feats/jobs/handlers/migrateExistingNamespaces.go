package handlers

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/k8s"
	"go.dfds.cloud/ssu-k8s/core/logging"
	selfserviceapi "go.dfds.cloud/ssu-k8s/core/ssu/selfservice-api"
	"go.dfds.cloud/ssu-k8s/feats/operator/misc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MigrateExistingNamespaces(ctx context.Context) error {
	logging.Logger.Info("Migrate existing namespaces")

	var namespacesCount int = 0
	var namespacesWithoutDfdsCapabilityLabelCount int = 0

	client, err := k8s.GetK8sClient()
	if err != nil {
		return err
	}

	nsListResp, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	namespacesWithoutDfdsCapabilityLabel := map[string]v1.Namespace{}
	updatedNamespaces := map[string]v1.Namespace{}

	for _, ns := range nsListResp.Items {
		if _, ok := ns.Labels[misc.LabelCapabilityKey]; !ok {
			namespacesWithoutDfdsCapabilityLabel[ns.Name] = ns
		}
	}

	namespacesCount = len(nsListResp.Items)
	namespacesWithoutDfdsCapabilityLabelCount = len(namespacesWithoutDfdsCapabilityLabel)

	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}
	ssApi := selfserviceapi.NewClient(conf.SelfserviceApi)
	capResp, err := ssApi.GetCapabilities()
	if err != nil {
		return err
	}

	for _, capa := range capResp {
		if ns, ok := namespacesWithoutDfdsCapabilityLabel[capa.RootID]; ok {
			if len(capa.Contexts) != 0 {
				ns.Labels[misc.LabelCapabilityKey] = capa.RootID
				ns.Labels[misc.LabelContextIdKey] = capa.Contexts[0].ID
				ns.Labels[misc.LabelAwsAccountKey] = capa.Contexts[0].AwsAccountID
				ns.Labels[misc.LabelReconcileKey] = "true"
				updatedNamespaces[ns.Name] = ns
			}
		}
	}

	for _, ns := range updatedNamespaces {
		fmt.Printf("Updating namespace %s\n", ns.Name)
		for k, v := range ns.Labels {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	fmt.Printf("namespaces count: %d\n", namespacesCount)
	fmt.Printf("namespaces without Dfds capability label: %d\n", namespacesWithoutDfdsCapabilityLabelCount)
	fmt.Printf("updated namespaces count: %d\n", len(updatedNamespaces))

	return nil
}
