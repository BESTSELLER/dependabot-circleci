package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
)

func CreateBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, baseBranch string, commitBranch *string) (bool, error) {
	// check if branch exists or if an older update exists
	branches, _, err := client.Repositories.ListBranches(ctx, repoOwner, repoName, nil)
	if err != nil {
		return false, err
	}

	branchComponent := strings.Split(*commitBranch, "@")

	for _, branch := range branches {
		branchName := branch.GetName()

		// exists ?
		if branchName != *commitBranch {
			return true, nil
		}

		// older update ?
		if strings.Contains(branchName, branchComponent[0]) && branchName != *commitBranch {
			// delete the branch
			fmt.Println("delete")
			_, err := client.Git.DeleteRef(ctx, repoOwner, repoName, "refs/heads/"+branchName)
			if err != nil {
				return false, err
			}
		}
	}

	var baseRef *github.Reference
	baseRef, _, err = client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+baseBranch)
	if err != nil {
		return false, err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + *commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	_, _, err = client.Git.CreateRef(ctx, repoOwner, repoName, newRef)
	if err != nil {
		return false, err
	}

	return false, nil

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
