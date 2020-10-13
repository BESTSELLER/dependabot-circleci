package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
)

// CheckPR
func CheckPR(ctx context.Context, client *github.Client, repoOwner string, repoName string, baseBranch string, commitBranch *string, commitMessage *string, component string) (bool, error) {
	pullreqs, _, _ := client.PullRequests.List(ctx, repoOwner, repoName, nil)
	for _, pr := range pullreqs {

		if pr.GetUser().GetLogin() != "dependabot-circleci[bot]" {
			continue
		}

		title := pr.GetTitle()

		// exists ?
		if title == *commitMessage {
			return true, nil
		}

		// older update ?
		if strings.Contains(title, fmt.Sprintf("@%s", component)) {

			branchURL, err := createBranch(ctx, client, repoOwner, repoName, baseBranch, commitBranch)
			if err != nil {
				return false, err
			}

			_, _, err = client.Issues.CreateComment(ctx, repoOwner, repoName, pr.GetNumber(), &github.IssueComment{Body: github.String(fmt.Sprintf("pullRequest superseeded by [%s](%s)", *commitMessage, branchURL))})
			if err != nil {
				return false, err
			}

			// delete old branch
			_, err = client.Git.DeleteRef(ctx, repoOwner, repoName, "refs/heads/"+pr.GetHead().GetRef())
			if err != nil {
				return false, err
			}

			return false, nil

		}

	}

	_, err := createBranch(ctx, client, repoOwner, repoName, baseBranch, commitBranch)
	if err != nil {
		return false, err
	}

	return false, nil
}

// createBranch creates a new commit branch for a specific update
func createBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, baseBranch string, commitBranch *string) (string, error) {

	var baseRef *github.Reference
	baseRef, _, err := client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+baseBranch)
	if err != nil {
		return "", err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + *commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	newBranch, _, err := client.Git.CreateRef(ctx, repoOwner, repoName, newRef)
	if err != nil {
		return "", err
	}

	return newBranch.GetURL(), nil

}

// UpdateFile updates the circleci config and creates a commit
func UpdateFile(ctx context.Context, client *github.Client, repoOwner string, repoName string, options *github.RepositoryContentFileOptions) error {
	_, _, err := client.Repositories.UpdateFile(ctx, repoOwner, repoName, ".circleci/config.yml", options)
	if err != nil {
		return err
	}
	return nil
}

// CreatePR creates a pull request with the new config
func CreatePR(ctx context.Context, client *github.Client, repoOwner string, repoName string, options *github.NewPullRequest) error {
	_, _, err := client.PullRequests.Create(ctx, repoOwner, repoName, options)
	if err != nil {
		return err
	}

	return nil
}
