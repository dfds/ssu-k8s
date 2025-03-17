package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Kubernetes struct {
		ClusterName string `json:"clusterName"`
	} `json:"kubernetes"`
	Git struct {
		Branch            string `json:"branch"`
		RemoteRepoURI     string `json:"remoteRepoUri"`
		TemporaryRepoPath string `json:"temporaryRepoPath"`
	}
	Enable struct {
		Messaging bool `json:"messaging" default:"true"`
		Operator  bool `json:"operator" default:"true"`
	} `json:"enable"`
}

const APP_CONF_PREFIX = "SSU_K8S"

func LoadConfig() (Config, error) {
	var conf Config
	err := envconfig.Process(APP_CONF_PREFIX, &conf)

	return conf, err
}
