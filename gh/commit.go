package gh

import (
	"context"

	"github.com/google/go-github/v32/github"
)

func CreateBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, baseBranch string, commitBranch *string) error {
	var baseRef *github.Reference

	baseRef, _, err := client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+baseBranch)
	if err != nil {
		return err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + *commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	_, _, err = client.Git.CreateRef(ctx, repoOwner, repoName, newRef)
	if err != nil {
		return err
	}

	return nil

}

func UpdateFile(ctx context.Context, client *github.Client, repoOwner string, repoName string, options *github.RepositoryContentFileOptions) error {
	_, _, err := client.Repositories.UpdateFile(ctx, repoOwner, repoName, ".circleci/config.yml", options)
	if err != nil {
		return err
	}
	return nil
}

func CreatePR(ctx context.Context, client *github.Client, repoOwner string, repoName string, options *github.NewPullRequest) error {
	_, _, err := client.PullRequests.Create(ctx, repoOwner, repoName, options)
	if err != nil {
		return err
	}

	return nil
}
