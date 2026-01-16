package selfservice_api

import (
	"encoding/json"
	"errors"
	"strings"
)

type GetCapabilitiesResponse struct {
	Items []*GetCapabilitiesResponseContextCapability `json:"items"`
}

func (g *GetCapabilitiesResponseContextCapability) GetContext() (*GetCapabilitiesResponseContext, error) {
	if len(g.Contexts) > 0 {
		if g.Contexts[0].AwsAccountID == "" {
			return g.Contexts[0], errors.New("capability has a Context, but no AWS account associated with the aforementioned Context")
		}
		return g.Contexts[0], nil
	} else {
		return nil, errors.New("capability doesn't have a Context")
	}
}

type GetCapabilitiesResponseContextCapability struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RootID      string `json:"rootId"`
	Description string `json:"description"`
	Metadata    string `json:"jsonMetadata"`
	Members     []struct {
		Email string `json:"email"`
	} `json:"members"`
	Contexts []*GetCapabilitiesResponseContext `json:"contexts,omitempty"`
}

func (g *GetCapabilitiesResponseContextCapability) HasMember(email string) bool {
	for _, member := range g.Members {
		if strings.ToLower(member.Email) == strings.ToLower(email) {
			return true
		}
	}

	return false
}

func (g *GetCapabilitiesResponseContextCapability) MetadataAsMap() (map[string]interface{}, error) {
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(g.Metadata), &payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

type GetCapabilitiesResponseContext struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AwsAccountID string `json:"awsAccountId"`
	AwsRoleArn   string `json:"awsRoleArn"`
	AwsRoleEmail string `json:"awsRoleEmail"`
}
