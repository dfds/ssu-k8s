package k8s

import (
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

type CachedResources struct {
	Deployments            *appsV1.DeploymentList
	Services               *coreV1.ServiceList
	DeploymentsByNamespace map[string]*appsV1.DeploymentList
	ServicesByNamespace    map[string]*coreV1.ServiceList
}

func NewCachedResources() *CachedResources {
	return &CachedResources{
		Deployments:            nil,
		Services:               nil,
		DeploymentsByNamespace: make(map[string]*appsV1.DeploymentList),
		ServicesByNamespace:    make(map[string]*coreV1.ServiceList),
	}
}
