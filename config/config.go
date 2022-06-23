package config

import (
	"strings"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type DatadogConfig struct {
	APIKey string `yaml:"api_key"`
}

type HTTPConfig struct {
	Token string `yaml:"token"`
}

// Config contains global config
type Config struct {
	Datadog DatadogConfig      `yaml:"datadog"`
	Github  githubapp.Config   `yaml:"github"`
	HTTP    HTTPConfig         `yaml:"http"`
	Server  baseapp.HTTPConfig `yaml:"server"`
}

// DBConfig contains global db config
type DBConfigSpec struct {
	ConnectionName string `yaml:"connection_name"`
	DBName         string `yaml:"db_name"`
	Password       string `yaml:"password"`
	Username       string `yaml:"username"`
}

// RepoConfig contains specific config for each repos
type RepoConfig struct {
	TargetBranch string   `yaml:"target-branch,omitempty"`
	Reviewers    []string `yaml:"reviewers,omitempty"`
	Assignees    []string `yaml:"assignees,omitempty"`
	Labels       []string `yaml:"labels,omitempty"`
	Directory    string   `yaml:"directory,omitempty"`
	Schedule     string   `yaml:"schedule,omitempty"`
}

// AppConfig contains global app config
var AppConfig Config

// DBConfig contains global db app config
var DBConfig DBConfigSpec

var Version = "1.0.0"

// ReadAppConfig reads a yaml config file
func ReadAppConfig(secrets []byte) error {
	if err := yaml.UnmarshalStrict(secrets, &AppConfig); err != nil {
		return errors.Wrap(err, "failed parsing configuration file")
	}

	return nil
}

// ReadDBConfig reads a yaml config file
func ReadDBConfig(secrets []byte) error {
	if err := yaml.UnmarshalStrict(secrets, &DBConfig); err != nil {
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

// IsValid checks if Repoconfig is valid
func (r RepoConfig) IsValid() error {
	var errMsg []string

	// check schedule
	switch strings.ToLower(r.Schedule) {
	case "daily", "weekly", "monthly", "":

	default:
		errMsg = append(errMsg, "invalid schedule")
	}

	if len(errMsg) != 0 {
		return errors.Errorf(strings.Join(errMsg, ", "))
	}
	return nil
}
