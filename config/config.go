package config

import (
	"io/ioutil"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config contains global config
type Config struct {
	Github githubapp.Config   `yaml:"github"`
	Server baseapp.HTTPConfig `yaml:"server"`
}

// RepoConfig contains specific config for each repos
type RepoConfig struct {
	TargetBranch string   `yaml:"target-branch,omitempty"`
	Reviewers    []string `yaml:"reviewers,omitempty"`
	Assignees    []string `yaml:"assignees,omitempty"`
	Labels       []string `yaml:"labels,omitempty"`
}

// ReadConfig reads a yaml config file
func ReadConfig(path string) (*Config, error) {
	var c Config

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.UnmarshalStrict(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}

	return &c, nil
}

// ReadRepoConfig reads a yaml file
func ReadRepoConfig(content []byte) (*RepoConfig, error) {

	var repoConfig RepoConfig

	if err := yaml.UnmarshalStrict(content, &repoConfig); err != nil {
		return nil, errors.Wrap(err, "failed parsing repository configuration file")
	}

	return &repoConfig, nil
}
