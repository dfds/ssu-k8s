package git

import (
	"errors"
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.uber.org/zap"
	"os"
)

type Repo struct {
	Path      string
	Branch    string
	RemoteURL string
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

	return &Repo{}, nil
}

func (repo *Repo) List() error {
	return nil
}

func (repo *Repo) Add() error {
	return nil
}

func (repo *Repo) Exists() error {
	return nil
}
