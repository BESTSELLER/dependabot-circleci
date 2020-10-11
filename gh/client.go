package gh

import (
	"context"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-githubapp/githubapp"
)

// GetOrganizationClient returns a github client
func GetOrganizationClient(ctx context.Context, config githubapp.Config, org string) (*github.Client, error) {

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
	install, err := installations.GetByOwner(ctx, org)
	if err != nil {
		return nil, err
	}

	// create a client to perform actions on that specific organization
	client, err := cc.NewInstallationClient(install.ID)
	return client, err
}

func createGHClient(config githubapp.Config) (githubapp.ClientCreator, error) {
	cc, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientUserAgent("dependabot-circleci/1.0.0"),
		githubapp.WithClientTimeout(3*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
	)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
