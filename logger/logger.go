package logger

import (
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/rs/zerolog"
)

// Init sets the global loglevel
func Init() {

	if config.EnvVars.LogLevel == nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}

	if *config.EnvVars.LogLevel <= int(zerolog.Disabled) {
		zerolog.SetGlobalLevel(zerolog.Level(*config.EnvVars.LogLevel))
	}
}
