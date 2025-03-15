package model

type Capability struct {
	Name         string `json:"name"`
	Id           string `json:"id"`
	ContextId    string `json:"contextId"`
	AwsAccountId string `json:"awsAccountId"`
}
