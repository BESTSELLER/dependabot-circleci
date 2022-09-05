package gh

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"

	conf "github.com/BESTSELLER/dependabot-circleci/config"
)

// GetOrganizationClients returns a github client
func GetOrganizationClients(config githubapp.Config) ([]*github.Client, error) {

	ctx := context.Background()

	cc, err := CreateGHClient(config)
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

func CreateGHClient(config githubapp.Config) (githubapp.ClientCreator, error) {
	cc, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientUserAgent(fmt.Sprintf("dependabot-circleci/%s", conf.Version)),
		githubapp.WithClientTimeout(10*time.Minute),
		githubapp.WithTransport(NewRateLimitTransport(http.DefaultTransport)),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
	)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

// GetSingleOrganizationClient returns a single github client
func GetSingleOrganizationClient(cc githubapp.ClientCreator, org string) (*github.Client, error) {
	ctx := context.Background()

	// create a client to perform actions as the application
	appClient, err := cc.NewAppClient()
	if err != nil {
		return nil, err
	}

	// look up the installation ID for a particular organization
	installations := githubapp.NewInstallationsService(appClient)
	install, err := installations.GetByOwner(ctx, org)
	if err != nil {
		return nil, err
	}

	// create a client to perform actions on that specific organization
	client, err := cc.NewInstallationClient(install.ID)
	return client, err
}
