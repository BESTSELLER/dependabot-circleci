package config

import (
	"io/ioutil"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type DatadogConfig struct {
	APIKey string `yaml:"api_key"`
}

// Config contains global config
type Config struct {
	Datadog DatadogConfig      `yaml:"datadog"`
	Github  githubapp.Config   `yaml:"github"`
	Server  baseapp.HTTPConfig `yaml:"server"`
}

// RepoConfig contains specific config for each repos
type RepoConfig struct {
	TargetBranch string   `yaml:"target-branch,omitempty"`
	Reviewers    []string `yaml:"reviewers,omitempty"`
	Assignees    []string `yaml:"assignees,omitempty"`
	Labels       []string `yaml:"labels,omitempty"`
}

// AppConfig contains global app config
var AppConfig Config

// ReadConfig reads a yaml config file
func ReadConfig(path string) error {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &AppConfig); err != nil {
		return errors.Wrap(err, "failed parsing configuration file")
	}

	return nil
}

// ReadRepoConfig reads a yaml file
func ReadRepoConfig(content []byte) (*RepoConfig, error) {

	var repoConfig RepoConfig

	if err := yaml.UnmarshalStrict(content, &repoConfig); err != nil {
		return nil, errors.Wrap(err, "failed parsing repository configuration file")
	}

	return &repoConfig, nil
}
