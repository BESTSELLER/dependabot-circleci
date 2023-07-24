package logger

import (
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init sets the global loglevel
func Init() {
	log.Debug().Msgf("Setting log level to %d (%s)", *config.EnvVars.LogLevel, zerolog.Level(*config.EnvVars.LogLevel).String())

	if config.AppConfig.LogLevel == nil && config.EnvVars.LogLevel == nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}

	if *config.EnvVars.LogLevel <= int(zerolog.Disabled) {
		zerolog.SetGlobalLevel(zerolog.Level(*config.EnvVars.LogLevel))
		return
	}

	if *config.AppConfig.LogLevel <= int(zerolog.Disabled) {
		zerolog.SetGlobalLevel(zerolog.Level(*config.AppConfig.LogLevel))
		return
	}
}
