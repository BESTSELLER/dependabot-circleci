package gh

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/rs/zerolog/log"
)

// CheckPR .
func CheckPR(ctx context.Context, client *github.Client, repoOwner string, repoName string, baseBranch string, commitBranch string, commitMessage string, component string) (bool, []*github.PullRequest, error) {
	PRsToBeClosed := []*github.PullRequest{}
	pullreqs, _, _ := client.PullRequests.List(ctx, repoOwner, repoName, nil)
	for _, pr := range pullreqs {

		if pr.GetUser().GetLogin() != "dependabot-circleci[bot]" {
			continue
		}

		title := pr.GetTitle()

		// exists ?
		if title == commitMessage {
			return true, nil, nil
		}

		// older update ?
		if strings.Contains(title, fmt.Sprintf("@%s", component)) {
			PRsToBeClosed = append(PRsToBeClosed, pr)
		}

	}
	if len(PRsToBeClosed) > 0 {
		return false, PRsToBeClosed, nil
	}

	return false, nil, nil
}

// CheckBranch checks if a branch already exists, in order to skip CreateBranch if needed
func CheckBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, commitBranch *string) bool {
	// Assumes that an error means that the branch do not exists
	_, _, err := client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+*commitBranch)
	if err != nil {
		return true
	} else {
		return false
	}
}

// CreateBranch creates a new commit branch for a specific update
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

// UpdateFile updates the circleci config and creates a commit
// we ignore conflicts as it saves us API calls and avoids GHs ratelimit
func UpdateFile(ctx context.Context, client *github.Client, repoOwner string, repoName string, file string, options *github.RepositoryContentFileOptions) error {
	_, resp, err := client.Repositories.UpdateFile(ctx, repoOwner, repoName, strings.TrimPrefix(file, "/"), options)
	if resp.StatusCode == http.StatusConflict {
		log.Debug().Str("repo_name", repoName).Err(err).Msg("Conflict, the updated file might already exists")
	}
	if err != nil && resp.StatusCode != http.StatusConflict {
		return err
	}
	return nil
}

// CreatePR creates a pull request with the new config
func CreatePR(ctx context.Context, client *github.Client, repoOwner string, repoName string, reviewers []string, assignees []string, labels []string, options *github.NewPullRequest) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Create(ctx, repoOwner, repoName, options)
	if err != nil {
		return nil, err
	}

	// disect team reviewers
	teamReviewers := []string{}
	singleReviewers := []string{}
	for _, reviewer := range reviewers {
		if strings.Contains(reviewer, "/") {
			teamReviewers = append(teamReviewers, strings.Split(reviewer, "/")[1])
			continue
		}
		singleReviewers = append(singleReviewers, reviewer)
	}

	log.Debug().Str("repo_name", repoName).Str("all_reviewers", fmt.Sprintf("%s", reviewers)).Msg("This is ALL the reviewers")
	log.Debug().Str("repo_name", repoName).Str("single_reviewers", fmt.Sprintf("%s", singleReviewers)).Msg("This is the SINGLE reviewers")
	log.Debug().Str("repo_name", repoName).Str("team_reviewers", fmt.Sprintf("%s", teamReviewers)).Msg("This is TEAM reviewers")

	// add default labels
	labels = append(labels, []string{"dependencies", "circleci"}...)

	// Add reviewers
	if len(singleReviewers) > 0 || len(teamReviewers) > 0 {
		_, _, err = client.PullRequests.RequestReviewers(ctx, repoOwner, repoName, pr.GetNumber(), github.ReviewersRequest{Reviewers: singleReviewers, TeamReviewers: teamReviewers})
		if err != nil {
			log.Error().Str("repo_name", repoName).Err(err).Msg("Failed to request reviewers")
		}
	}

	// Add asignees
	_, _, err = client.Issues.AddAssignees(ctx, repoOwner, repoName, pr.GetNumber(), assignees)
	if err != nil {
		log.Error().Str("repo_name", repoName).Err(err).Msg("Failed to add assignees")
	}

	// Add labels
	_, _, err = client.Issues.AddLabelsToIssue(ctx, repoOwner, repoName, pr.GetNumber(), labels)
	if err != nil {
		log.Error().Str("repo_name", repoName).Err(err).Msg("Failed to add labels")
	}

	return pr, nil
}

// CleanUpOldBranch comments the old pull request and deletes the old branch
func CleanUpOldBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, PRs []*github.PullRequest, newPR int) {

	for _, pr := range PRs {
		prNumber := pr.GetNumber()
		_, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, prNumber, &github.IssueComment{Body: github.String(fmt.Sprintf("Update superseded by #%d", newPR))})
		if err != nil {
			log.Error().Err(err).Msgf("could not create a comment on #%d in repo '%s'", prNumber, repoName)
			continue
		}

		// make sure the old pull request is closed
		_, _, err = client.PullRequests.Edit(ctx, repoOwner, repoName, prNumber, &github.PullRequest{State: github.String("closed")})
		if err != nil {
			log.Error().Err(err).Msgf("could not delete pr #%d on '%s'", prNumber, repoName)
			continue
		}

		// delete old branch
		ref := "refs/heads/" + pr.GetHead().GetRef()
		_, err = client.Git.DeleteRef(ctx, repoOwner, repoName, ref)
		if err != nil {
			log.Debug().Err(err).Msgf("could not delete ref '%s' from repo '%s'", ref, repoName)
			continue
		}

	}

}
