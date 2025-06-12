package handlers

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/k8s"
	"go.dfds.cloud/ssu-k8s/core/logging"
	v1Apps "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func CacheKubernetesResources(ctx context.Context) error {
	logging.Logger.Info("Caching Kubernetes resources for API")
	client, err := k8s.GetDynamicK8sClient()
	if err != nil {
		return err
	}

	resources, err := getK8sResources(client, ctx, "", "deployments", "apps", "v1")
	if err != nil {
		return err
	}

	for _, res := range resources.Items {
		logging.Logger.Info(fmt.Sprintf("%s: %s - %s", res.GetObjectKind().GroupVersionKind().Kind, res.GetNamespace(), res.GetName()))

		if res.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
			deployment, err := mapFromUnstructured[v1Apps.Deployment](res)
			if err != nil {
				return err
			}

			logging.Logger.Info(fmt.Sprintf("  Replicas: %d", *deployment.Spec.Replicas))
		}
	}

	return nil
}

func getK8sResources(client *dynamic.DynamicClient, ctx context.Context, ns string, kind string, apiGroup string, version string) (*unstructured.UnstructuredList, error) {
	resources, err := client.Resource(schema.GroupVersionResource{
		Group:    apiGroup,
		Resource: kind,
		Version:  version,
	}).List(ctx, v1.ListOptions{})

	return resources, err
}

func mapFromUnstructured[T any](item unstructured.Unstructured) (T, error) {
	var converted T
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &converted)
	return converted, err
}
