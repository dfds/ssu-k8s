package handlers

import (
	"context"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/k8s"
	"go.dfds.cloud/ssu-k8s/core/logging"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func CacheKubernetesResources(ctx context.Context) error {
	logging.Logger.Info("Caching Kubernetes resources for API")

	cache := k8s.NewCachedResources()

	client, err := k8s.GetK8sClient()
	if err != nil {
		return err
	}

	dynClient, err := k8s.GetDynamicK8sClient()
	if err != nil {
		return err
	}

	deployments, err := client.AppsV1().Deployments("").List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}
	cache.Deployments = deployments

	services, err := client.CoreV1().Services("").List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}
	cache.Services = services

	ingressroutesDyn, err := dynClient.Resource(schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutes",
	}).List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	ingressRoutes := []k8s.IngressRoute{}

	routesWithPathPrefix := 0

	for _, ingr := range ingressroutesDyn.Items {
		ingressRoute, err := k8s.MapFromUnstructured[k8s.IngressRoute](ingr)
		if err != nil {
			return err
		}

		hasPathPrefix := false

		ingressRoute.PopulateDefaultsIfEmpty()

		for _, route := range ingressRoute.Spec.Routes {
			logging.Logger.Info(route.Match)
			extractedRule, err := route.ParseMatch()
			if err != nil {
				return err
			}

			fmt.Printf("  Host: %s\n", extractedRule.Host)
			fmt.Printf("  Pathprefix: %s\n", extractedRule.PathPrefix)

			for _, svc := range route.Services {
				fmt.Printf("  Namespace: %s\n", svc.Namespace)
				fmt.Printf("  Service: %s\n", svc.Name)
				fmt.Printf("  Port: %s\n", svc.GetPort())
			}

			if extractedRule.PathPrefix != "" {
				hasPathPrefix = true
			}
		}

		if hasPathPrefix {
			routesWithPathPrefix = routesWithPathPrefix + 1
		}

		ingressRoutes = append(ingressRoutes, ingressRoute)
	}

	logging.Logger.Info(fmt.Sprintf("Deployments: %d", len(deployments.Items)))
	logging.Logger.Info(fmt.Sprintf("Services: %d", len(services.Items)))
	logging.Logger.Info(fmt.Sprintf("IngressRoutes: %d", len(ingressRoutes)))
	logging.Logger.Info(fmt.Sprintf("IngressRoutes with PathPrefix: %d", routesWithPathPrefix))

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
