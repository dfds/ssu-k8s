package git

import (
	"bytes"
	"errors"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/operator/model"
	"go.uber.org/zap"
	"os"
	"text/template"
)

type Repo struct {
	config Config
	logger *zap.Logger
}

type Config struct {
	RemoteRepoUri     string
	TemporaryRepoPath string
	GitUsername       string
	GitEmail          string
	Branch            string
}

func LoadRepo(conf Config) (*Repo, error) {
	logger := logging.Logger.With(zap.String("repo", conf.RemoteRepoUri), zap.String("branch", conf.Branch), zap.String("temporary_repo_path", conf.TemporaryRepoPath))

	if conf.RemoteRepoUri == "" || conf.TemporaryRepoPath == "" {
		return nil, errors.New("LoadRepo: remote repo URI or temporary repo path is empty")
	}

	if _, err := os.Stat(conf.TemporaryRepoPath); os.IsNotExist(err) {
		err = os.MkdirAll(conf.TemporaryRepoPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/.git", conf.TemporaryRepoPath)); os.IsNotExist(err) {
		resp, err := ExecuteCmd("git", conf.TemporaryRepoPath, []string{"clone", conf.RemoteRepoUri, "."})
		if err != nil {
			logger.Debug(resp)
			return nil, err
		}
		logger.Debug(resp)

		resp, err = ExecuteCmd("git", conf.TemporaryRepoPath, []string{"checkout", conf.Branch})
		if err != nil {
			logger.Debug(resp)
			return nil, err
		}
		logger.Debug(resp)
	} else {
		resp, err := ExecuteCmd("git", conf.TemporaryRepoPath, []string{"fetch", "origin"})
		if err != nil {
			logger.Debug(resp)
			return nil, err
		}
		logger.Debug(resp)

		resp, err = ExecuteCmd("git", conf.TemporaryRepoPath, []string{"reset", "--hard", fmt.Sprintf("origin/%s", conf.Branch)})
		if err != nil {
			logger.Debug(resp)
			return nil, err
		}
		logger.Debug(resp)
	}

	return &Repo{
		config: conf,
		logger: logger,
	}, nil
}

func (repo *Repo) List() error {
	return nil
}

const capabilityBaseFileName = "capability-base.yaml"
const kustomizationFileName = "kustomization.yaml"

func (repo *Repo) Add(capability model.Capability, clusterName string) error {
	logger := logging.Logger.With(zap.String("cluster_name", clusterName), zap.String("capabilityId", capability.Id))
	capabilityManifestPath := capabilityManifestPath(repo.config.TemporaryRepoPath, clusterName, capability)

	if _, err := os.Stat(capabilityManifestPath); os.IsNotExist(err) {
		err = os.MkdirAll(capabilityManifestPath, 0755)
		if err != nil {
			return err
		}
	}

	vars := map[string]interface{}{
		"capabilityName": capability.Name,
		"capabilityId":   capability.Id,
	}

	labels := map[string]string{
		"dfds.cloud/capability":  capability.Id,
		"dfds.cloud/reconcile":   "true",
		"dfds.cloud/context-id":  capability.ContextId,
		"dfds.cloud/aws-account": capability.AwsAccountId,
	}

	tmpVars := templateVars{
		Vars:   vars,
		Labels: labels,
	}

	if checkFileExists(fmt.Sprintf("%s/%s", capabilityManifestPath, kustomizationFileName)) {
		logger.Debug("Kustomization file already exists, skipping creation")
	} else {
		file, err := os.Create(fmt.Sprintf("%s/%s", capabilityManifestPath, kustomizationFileName))
		if err != nil {
			return err
		}
		file.Close()

		err = writeTemplate("kustomization.tpl", capabilityManifestPath, kustomizationFileName, tmpVars)
		if err != nil {
			return err
		}
	}

	if checkFileExists(fmt.Sprintf("%s/%s", capabilityManifestPath, capabilityBaseFileName)) {
		logger.Debug("Capability base file already exists, skipping creation")
	} else {
		file, err := os.Create(fmt.Sprintf("%s/%s", capabilityManifestPath, capabilityBaseFileName))
		if err != nil {
			return err
		}
		file.Close()

		err = writeTemplate("capability-base.tpl", capabilityManifestPath, capabilityBaseFileName, tmpVars)
		if err != nil {
			return err
		}
	}

	resp, err := ExecuteCmd("git", repo.config.TemporaryRepoPath, []string{"add", capabilityManifestPath})
	if err != nil {
		logger.Debug(resp)
		return err
	}
	logger.Debug(resp)

	msg := fmt.Sprintf("Capability changes for '%s'", capability.Id)
	author := fmt.Sprintf("\"ssu-k8s <ssu-k8s@dfds.cloud>\"")

	resp, err = ExecuteCmd("git", repo.config.TemporaryRepoPath, []string{"commit", "-m", msg, "--author", author})
	if err != nil {
		logger.Debug(resp)
		return err
	}
	logger.Debug(resp)

	resp, err = ExecuteCmd("git", repo.config.TemporaryRepoPath, []string{"push"})
	if err != nil {
		logger.Debug(resp)
		return err
	}
	logger.Debug(resp)

	return nil
}

func (repo *Repo) Exists() error {
	return nil
}

func capabilityManifestPath(repoPath string, clusterName string, capability model.Capability) string {
	return fmt.Sprintf("%s/cluster/%s/%s", repoPath, clusterName, capability.Id)
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

type templateVars struct {
	Vars   map[string]interface{}
	Labels map[string]string
}

func loadTemplate(name string) (*template.Template, error) {
	buf, err := os.ReadFile(fmt.Sprintf("core/git/template/%s", name))
	if err != nil {
		return nil, err
	}

	templateContainer := template.New(name)
	templateParsed, err := templateContainer.Parse(string(buf))
	if err != nil {
		return templateParsed, errors.New("unable to parse template file")
	}

	return templateParsed, nil
}

func writeTemplate(name string, path string, outputName string, vars templateVars) error {
	capabilityBaseTemplate, err := loadTemplate(name)
	if err != nil {
		return err
	}
	var body bytes.Buffer

	err = capabilityBaseTemplate.Execute(&body, vars)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", path, outputName), body.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
