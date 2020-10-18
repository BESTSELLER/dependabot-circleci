package main

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/logger"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var ctx = context.Background()
var org = os.Getenv("DEPENDABOT_ORG")

func main() {
	logger.Init()

	appConfig, err := config.ReadConfig(os.Getenv("DEPENDABOT_CONFIG"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config")
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, appConfig.Github, org)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	// create statsd client
	err = datadog.CreateClient()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register dogstatsd client")
	}

	dependabot.Start(ctx, client)
}
