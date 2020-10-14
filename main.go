package main

import (
	"context"
	"os"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var ctx = context.Background()
var org = os.Getenv("DEPENDABOT_ORG")

func main() {
	appConfig, err := config.ReadConfig("/secrets/secrets.json")
	if err != nil {
		panic(err)
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, appConfig.Github, org)
	if err != nil {
		panic(err)
	}

	dependabot.Start(ctx, client)
}
