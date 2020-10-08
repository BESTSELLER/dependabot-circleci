package gh

import (
	"context"

	"github.com/google/go-github/v32/github"
)

// GetRepos
func GetRepos(ctx context.Context, client *github.Client, org string) ([]*github.Repository, error) {
	repos, _, err := client.Repositories.List(ctx, org, nil)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func GetRepoContent(ctx context.Context, client *github.Client, repo *github.Repository) ([]byte, error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), ".circleci/config.yml", nil)
	if err != nil {
		return nil, err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}
