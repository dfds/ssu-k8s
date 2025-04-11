package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"go.dfds.cloud/ssu-k8s/core/config"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
	"text/template"

	"go.dfds.cloud/ssu-k8s/core/git"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const capabilityDataNamespace = "capability-data"

func ReconcileCapabilityResources(ctx context.Context, client client.Client, capability model.Capability, ns string, repo *git.Repo) error {
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	err = repo.Add(model.Capability{
		Name: capability.Name,
		Id:   capability.Id,
	}, conf.Kubernetes.ClusterName)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func ReconcileCapabilityDeploymentToken(ctx context.Context, client client.Client, capability model.Capability, ns string, secret *v1.Secret) error {
	// Assume AWS role in Capability account
	creds, err := AssumeRole(ctx, fmt.Sprintf("arn:aws:iam::%s:role/ssu-ssm", capability.AwsAccountId))
	if err != nil {
		return err
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken)), awsConfig.WithRegion("eu-central-1"))
	if err != nil {
		return err
	}

	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	parameterExists := true

	ssmClient := ssm.NewFromConfig(cfg)
	resp, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(fmt.Sprintf("/managed/ssu/k8s-deployment-%s", conf.Kubernetes.ClusterName)),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		paramNotFoundErr := &ssmTypes.ParameterNotFound{}
		if errors.As(err, &paramNotFoundErr) {
			parameterExists = false
		} else {
			return err
		}
	}

	kubeConfig, err := writeTemplate("kubeconfig", templateVars{
		Vars: map[string]interface{}{
			"clusterName": conf.Kubernetes.ClusterName,
			"namespace":   ns,
			"token":       string(secret.Data["token"]),
			"caData":      conf.Kubernetes.ClusterCa,
			"server":      conf.Kubernetes.ClusterEndpoint,
		},
		Labels: nil,
	})
	if err != nil {
		return err
	}

	if resp != nil && parameterExists {
		fmt.Println("Parameter found!")
		currentParameter := *resp.Parameter.Value
		if !strings.EqualFold(currentParameter, kubeConfig) {
			fmt.Println("Parameter out of date, updating")
			_, err = ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
				Name:      aws.String(fmt.Sprintf("/managed/ssu/k8s-deployment-%s", conf.Kubernetes.ClusterName)),
				Value:     aws.String(kubeConfig),
				DataType:  aws.String("text"),
				Overwrite: aws.Bool(true),
				Type:      "SecureString",
			})
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Println("Parameter not found!")
		_, err = ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
			Name:      aws.String(fmt.Sprintf("/managed/ssu/k8s-deployment-%s", conf.Kubernetes.ClusterName)),
			Value:     aws.String(kubeConfig),
			DataType:  aws.String("text"),
			Overwrite: aws.Bool(true),
			Type:      "SecureString",
		})
		if err != nil {
			return err
		}
	}

	// Check if SSM parameter exists in Capability AWS account, if not, create it, if it does, make sure it is up to date

	return nil
}

// CreateHttpClientWithoutKeepAlive Currently the AWS SDK seems to let connections live for way too long. On OSes that has a very low file descriptior limit this becomes an issue.
func CreateHttpClientWithoutKeepAlive() *awsHttp.BuildableClient {
	client := awsHttp.NewBuildableClient().WithTransportOptions(func(transport *http.Transport) {
		transport.DisableKeepAlives = true
	})

	return client
}

func AssumeRole(ctx context.Context, roleArn string) (*types.Credentials, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion("eu-west-1"), awsConfig.WithHTTPClient(CreateHttpClientWithoutKeepAlive()))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	roleSessionName := "ssu-k8s"

	assumedRole, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{RoleArn: &roleArn, RoleSessionName: &roleSessionName})
	if err != nil {
		log.Printf("unable to assume role %s, %v", roleArn, err)
		return nil, err
	}

	return assumedRole.Credentials, nil
}

type templateVars struct {
	Vars   map[string]interface{}
	Labels map[string]string
}

func writeTemplate(name string, vars templateVars) (string, error) {
	templateContainer := template.New(name)
	capabilityBaseTemplate, err := templateContainer.Parse(kubeconfigTemplate)
	if err != nil {
		errors.New("unable to parse template file")
	}

	var body bytes.Buffer

	err = capabilityBaseTemplate.Execute(&body, vars)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}

var kubeconfigTemplate string = `
apiVersion: v1
clusters:
- cluster:
    server: {{index .Vars "server"}}
    certificate-authority-data: {{index .Vars "caData"}}
  name: {{index .Vars "clusterName"}}
contexts:
- context:
    cluster: {{index .Vars "clusterName"}}
    user: deploy-user
    namespace: {{index .Vars "namespace"}}
  name: aws
current-context: aws
kind: Config
preferences: {}
users:
- name: deploy-user
  user:
    token: {{index .Vars "token"}}
`
