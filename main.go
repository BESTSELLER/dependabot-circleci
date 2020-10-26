package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/logger"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var ctx = context.Background()

func main() {
	err := config.LoadEnvConfig()
	logger.Init()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}

	appConfig, err := config.ReadConfig(config.EnvVars.Config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config")
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, appConfig.Github, config.EnvVars.Org)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	// create statsd client
	err = datadog.CreateClient()
	if err != nil {
		log.Error().Err(err).Msg("failed to register dogstatsd client")
	}

	dependabot.Start(ctx, client)
}
