package gh

import (
	"context"

	"github.com/google/go-github/v32/github"
)

// GetRepos returns a list of repos for an orginasation
func GetRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {

	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// GetRepoContent returns the circleci config as a byte array
func GetRepoContent(ctx context.Context, client *github.Client, repo *github.Repository, file string) ([]byte, *string, error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, repo.GetOwner().GetLogin(), repo.GetName(), file, nil)
	if err != nil {
		return nil, nil, err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, nil, err
	}

	return []byte(content), fileContent.SHA, nil
}
