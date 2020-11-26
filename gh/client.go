package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
)

var version string

// GetOrganizationClients returns a github client
func GetOrganizationClients(ctx context.Context, config githubapp.Config) ([]*github.Client, error) {

	cc, err := createGHClient(config)
	if err != nil {
		return nil, err
	}

	// create a client to perform actions as the application
	appClient, err := cc.NewAppClient()
	if err != nil {
		return nil, err
	}

	// look up the installation ID for a particular organization
	installations := githubapp.NewInstallationsService(appClient)

	// get organisations
	orgs, err := installations.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var clients []*github.Client
	for _, org := range orgs {
		install, err := installations.GetByOwner(ctx, org.Owner)
		if err != nil {
			return nil, err
		}

		client, err := cc.NewInstallationClient(install.ID)
		clients = append(clients, client)
	}

	// how do we handle errors here ?
	return clients, err
}

func createGHClient(config githubapp.Config) (githubapp.ClientCreator, error) {
	cc, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientUserAgent(fmt.Sprintf("dependabot-circleci/%s", version)),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
	)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
