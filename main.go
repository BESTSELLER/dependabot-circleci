package main

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/logger"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var ctx = context.Background()
var wg sync.WaitGroup

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

	// create statsd client
	err = datadog.CreateClient()
	if err != nil {
		log.Error().Err(err).Msg("failed to register dogstatsd client")
	}

	// create clients
	clients, err := gh.GetOrganizationClients(ctx, appConfig.Github)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	// send stats to dd
	go datadog.Gauge("organizations", float64(len(clients)), nil)

	// magic will happen
	for _, client := range clients {
		wg.Add(1)
		client := client
		go func() {
			defer wg.Done()
			dependabot.Start(ctx, client)
		}()
	}
	wg.Wait()

}
