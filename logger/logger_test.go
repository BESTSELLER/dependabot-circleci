package logger_test

import (
	"testing"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestInit tests the Init function
func TestInit(t *testing.T) {
	// Test case 1: both AppConfig and EnvVars are nil
	config.AppConfig.LogLevel = nil
	config.EnvVars.LogLevel = nil
	logger.Init()
	assert.Equal(t, zerolog.InfoLevel, logger.LogLevel)

	// Test case 2: EnvVars is set to Disabled
	config.AppConfig.LogLevel = nil
	config.EnvVars.LogLevel = new(int)
	*config.EnvVars.LogLevel = int(zerolog.Disabled)
	logger.Init()
	assert.Equal(t, zerolog.Disabled, logger.LogLevel)

	// Test case 3: AppConfig is set to Error
	config.AppConfig.LogLevel = new(int)
	*config.AppConfig.LogLevel = int(zerolog.ErrorLevel)
	config.EnvVars.LogLevel = nil
	logger.Init()
	assert.Equal(t, zerolog.ErrorLevel, logger.LogLevel)
}
