package gh

import (
	"context"

	"github.com/google/go-github/v39/github"
)

// GetRepos returns a list of repos for an orginasation
func GetRepos(ctx context.Context, client *github.Client, page int) ([]*github.Repository, error) {

	listRepos, resp, err := client.Apps.ListRepos(ctx, &github.ListOptions{PerPage: 100, Page: page})
	if err != nil {
		return nil, err
	}

	repos := listRepos.Repositories

	if resp.NextPage != 0 {
		moreRepos, _ := GetRepos(ctx, client, page+1)
		repos = append(moreRepos, repos...)
	}

	return repos, nil
}

// GetRepoContent returns the circleci config as a byte array
func GetRepoContent(ctx context.Context, client *github.Client, owner string, repo string, file string, branch string) ([]byte, *string, error) {
	options := &github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := client.Repositories.GetContents(context.Background(), owner, repo, file, options)
	if err != nil {
		return nil, nil, err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, nil, err
	}

	return []byte(content), fileContent.SHA, nil
}
