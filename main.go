package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/api"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/logger"

	"flag"
)

func init() {
	// set initial loglevel
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	err := config.LoadEnvConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}

	var appsecret []byte
	var dbsecret []byte

	log.Debug().Msgf("Using config file: %s", config.EnvVars.Config)
	bytes, err := os.ReadFile(config.EnvVars.Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read file %s", config.EnvVars.Config)
	}
	appsecret = bytes

	log.Debug().Msgf("Using db config file: %s", config.EnvVars.DBConfig)
	bytes, err = os.ReadFile(config.EnvVars.DBConfig)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to read file %s", config.EnvVars.DBConfig)
	}
	dbsecret = bytes

	err = config.ReadAppConfig(appsecret)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config:")
	}

	err = config.ReadDBConfig(dbsecret)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read db config:")
	}

	// init logger with proper loglevel from config
	logger.Init()
	log.Debug().Msgf("Logging level: %s", logger.LogLevel.String())
}

func main() {

	webhookFlag := flag.Bool("webhook", false, "Will start the webhook server.")
	workerFlag := flag.Bool("worker", false, "Will start the worker server.")
	controllerFlag := flag.Bool("controller", false, "Will start the controller.")

	flag.Parse()

	api.SetupRouter(*webhookFlag, *workerFlag, *controllerFlag)
}
