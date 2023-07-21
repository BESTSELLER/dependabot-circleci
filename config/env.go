package config

import (
	"github.com/kelseyhightower/envconfig"
)

// EnvConfig defines the structure of the global configuration parameters
type EnvConfig struct {
	Config    string `required:"false"`
	DBConfig  string `required:"false"`
	LogLevel  *int   `required:"false"`
	DDAddress string `required:"false"`
	Schedule  string `required:"false"`
	WorkerURL string `required:"false"`
}

// EnvVars stores the Global Configuration.
var EnvVars EnvConfig

// LoadEnvConfig Loads config from env
func LoadEnvConfig() error {
	err := envconfig.Process("dependabot", &EnvVars)
	if err != nil {
		return err
	}

	// If no log level is set, default to info
	if EnvVars.LogLevel == nil {
		logLevel := 1
		EnvVars.LogLevel = &logLevel
	}

	return nil
}
