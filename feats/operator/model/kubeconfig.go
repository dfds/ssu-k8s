package model

import (
	"bytes"
	"fmt"
	"text/template"
)

var kubeconfigTmpl = fmt.Sprintf(`
apiVersion: v1
clusters:
- cluster:
    server: {{.Endpoint}}
    certificate-authority-data: {{.CaCert}}
  name: {{.Name}}
contexts:
- context:
    cluster: {{.Name}}
    user: deploy-user
    namespace: {{.Namespace}}
  name: aws
current-context: aws
kind: Config
preferences: {}
users:
- name: deploy-user
  user:
    token: {{.Token}}
`)

type KubeConfigData struct {
	Name      string
	CaCert    string
	Endpoint  string
	Token     string
	Namespace string
}

func GenerateKubeConfig(data KubeConfigData) ([]byte, error) {
	kubeTemplate := template.Must(template.New("kubeconfig").Parse(kubeconfigTmpl))
	var buf bytes.Buffer
	err := kubeTemplate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
