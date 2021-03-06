package config

import (
	"github.com/kelseyhightower/envconfig"
)

// EnvConfig defines the structure of the global configuration parameters
type EnvConfig struct {
	Config    string `required:"true"`
	LogLevel  string `required:"false"`
	DDAddress string `required:"false"`
	Version   string `required:"false"`
	Schedule  string `required:"false"`
}

// EnvVars stores the Global Configuration.
var EnvVars EnvConfig

//LoadEnvConfig Loads config from env
func LoadEnvConfig() error {
	err := envconfig.Process("dependabot", &EnvVars)
	if err != nil {
		return err
	}
	return nil
}
