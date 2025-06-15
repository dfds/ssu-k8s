package events

type K8sNamespaceCreatedAndAwsArnConnected struct {
	CapabilityId  string `json:"capabilityId"`
	ContextId     string `json:"contextId"`
	NamespaceName string `json:"namespaceName"`
}
