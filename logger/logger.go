package logger

import (
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/rs/zerolog"
)

var LogLevel zerolog.Level

// Init sets the global loglevel
func Init() {
	// sets default to info if nothing is set
	if config.AppConfig.LogLevel == nil && config.EnvVars.LogLevel == nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		LogLevel = zerolog.InfoLevel
		return
	}

	// env is prioritized over config, as this has less dependencies
	if config.EnvVars.LogLevel != nil && *config.EnvVars.LogLevel <= int(zerolog.Disabled) {
		zerolog.SetGlobalLevel(zerolog.Level(*config.EnvVars.LogLevel))
		LogLevel = zerolog.Level(*config.EnvVars.LogLevel)
		return
	}

	// if not set in env, set from config
	if *config.AppConfig.LogLevel <= int(zerolog.Disabled) {
		zerolog.SetGlobalLevel(zerolog.Level(*config.AppConfig.LogLevel))
		LogLevel = zerolog.Level(*config.AppConfig.LogLevel)
		return
	}
}
