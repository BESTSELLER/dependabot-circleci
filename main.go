package main

import (
	"context"
	"log"
	"os"

	"github.com/BESTSELLER/dependabot-circleci/datadog"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var ctx = context.Background()
var org = os.Getenv("DEPENDABOT_ORG")

func main() {
	appConfig, err := config.ReadConfig(os.Getenv("DEPENDABOT_CONFIG"))
	if err != nil {
		panic(err)
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, appConfig.Github, org)
	if err != nil {
		panic(err)
	}

	// create statsd client
	err = datadog.CreateClient()
	if err != nil {
		log.Fatalf("failed to register dogstatsd client: %v \n", err)
	}

	dependabot.Start(ctx, client)
}
