package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/CircleCI-Public/circleci-cli/api"
	"github.com/CircleCI-Public/circleci-cli/api/graphql"
	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

func main() {
	config, err := config.ReadConfig("./config.yml")
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	server, err := baseapp.NewServer(
		config.Server,
		baseapp.DefaultParams(logger, "exampleapp.")...,
	)
	if err != nil {
		panic(err)
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.Github,
		githubapp.WithClientUserAgent("example-app/1.0.0"),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(server.Registry()),
		),
	)
	if err != nil {
		panic(err)
	}

	ctx, client, err := getOrganizationClient(cc, "brondum")
	if err != nil {
		panic(err)
	}

	repos, _, err := client.Repositories.List(ctx, "brondum", nil)
	if err != nil {
		panic(err)
	}

	for _, v := range repos {
		fmt.Println(v.GetName())

		fileContent, _, _, err := client.Repositories.GetContents(ctx, v.GetOwner().GetLogin(), v.GetName(), ".circleci/config.yml", nil)
		if err != nil {
			continue
		}

		content, _ := fileContent.GetContent()

		var some yaml.Node
		err = yaml.Unmarshal([]byte(content), &some)
		if err != nil {
			panic(err)
		}

		rabbitHole(&some)

		by, err := yaml.Marshal(&some)
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("output.yml", by, 0644)
		if err != nil {
			panic(err)
		}
	}

	// Start is blocking
	err = server.Start()
	if err != nil {
		panic(err)
	}
}

func getOrganizationClient(cc githubapp.ClientCreator, org string) (context.Context, *github.Client, error) {

	ctx := context.Background()

	// create a client to perform actions as the application
	appClient, err := cc.NewAppClient()
	if err != nil {
		return ctx, nil, err
	}

	// look up the installation ID for a particular organization
	installations := githubapp.NewInstallationsService(appClient)
	install, err := installations.GetByOwner(ctx, org)
	if err != nil {
		return ctx, nil, err
	}

	// create a client to perform actions on that specific organization
	client, err := cc.NewInstallationClient(install.ID)
	return ctx, client, err
}

func rabbitHole(node *yaml.Node) {
	for i, nextHole := range node.Content {
		if nextHole.Value == "orbs" {
			orbs := node.Content[i+1]
			extractOrbs(orbs.Content)
		}

		if nextHole.Value == "executors" {
			orbs := node.Content[i+1]
			extractImages(orbs.Content)
		}
		if nextHole.Value == "jobs" {
			orbs := node.Content[i+1]
			extractImages(orbs.Content)
		}

		rabbitHole(nextHole)
	}
}

func extractOrbs(orbs []*yaml.Node) {
	for i := 0; i < len(orbs); i = i + 2 {
		orb := orbs[i+1]
		orb.Value = findNewestOrbVersion(orb.Value)
	}
}

func extractImages(orbs []*yaml.Node) {
	for i := 0; i < len(orbs); i++ {
		orb := orbs[i]
		if orb.Value == "image" {
			orb = orbs[i+1]

			orb.Value = findNewestDockerVersion(orb.Value)
		}
		extractImages(orb.Content)
	}
}

func findNewestOrbVersion(orb string) string {

	orbSplitString := strings.Split(orb, "@")

	// check if orb is always updated
	if orbSplitString[1] == "volatile" {
		return "volatile"
	}

	client := graphql.NewClient("https://circleci.com/", "graphql-unstable", "", false)

	// if requests fails, return current version
	orbInfo, err := api.OrbInfo(client, orbSplitString[0])
	if err != nil {
		log.Printf("finding latests orb version failed: %v", err)
		return orbSplitString[1]
	}

	return orbInfo.Orb.HighestVersion
}

func findNewestDockerVersion(currentVersion string) string {

	current := strings.Split(currentVersion, ":")

	fmt.Println(current)
	fmt.Println(len(current))

	// check if image has no version tag
	if len(current) == 1 {
		return currentVersion
	}

	// check if tag is latest
	if strings.ToLower(current[1]) == "latest" {
		return currentVersion
	}

	// define registry
	registry := "registry.hub.docker.com"
	registryImage := strings.Split(current[0], "/")

	if len(registryImage) > 1 {
		registry = fmt.Sprintf("%s", strings.Join(registryImage[:len(registryImage)-1], "/"))
	}

	fmt.Println(registry)

	return "latest"
	// This one is a bit tricky actually! Watchtower seems to do this by utilising a docker client, but then we need
	// Docker in docker i guess ? Maybe there is a smart api endpoint, all registries should use the same to communicate with docker i guess ?
}
