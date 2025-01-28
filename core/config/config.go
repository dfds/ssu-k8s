package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Aws struct {
		AccountNamePrefix string `json:"accountNamePrefix"`
		AssumableRoles    struct {
			SsoManagementArn          string `json:"ssoManagementArn"`
			CapabilityAccountRoleName string `json:"capabilityAccountRoleName"`
		} `json:"assumableRoles"`
		OrganizationsParentId     string `json:"organizationsParentId"`
		RootOrganizationsParentId string `json:"rootOrganizationsParentId"`
	} `json:"aws"`
	Selfservice struct { // Capability-Service
		Host         string `json:"host"`
		TokenScope   string `json:"tokenScope"`
		ClientId     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	} `json:"capSvc"`
	Log struct {
		Level string `json:"level"`
		Debug bool   `json:"debug"`
	}
	EventHandling struct {
		Enabled bool `json:"enable"`
	}
	Http struct {
		Enabled bool `json:"enable"`
	}
	Metrics struct {
		Enabled bool `json:"enable"`
	}
	Kubernetes struct {
		ClusterName     string `json:"clusterName"`
		ClusterEndpoint string `json:"clusterEndpoint"`
		ClusterCaCert   string `json:"clusterCaCert"`
	}
}

const APP_CONF_PREFIX = "SSU_K8S"

func LoadConfig() (Config, error) {
	var conf Config
	err := envconfig.Process(APP_CONF_PREFIX, &conf)

	return conf, err
}
