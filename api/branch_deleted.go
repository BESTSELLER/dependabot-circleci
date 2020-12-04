package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v32/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
)

// BranchDeleteHandler handles all comments on issues
type BranchDeleteHandler struct {
	githubapp.ClientCreator
}

// Handles return list of events to listens to
func (h *BranchDeleteHandler) Handles() []string {
	return []string{"push"}
}

// Handle has ALL the logic! ;)
func (h *BranchDeleteHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse push event")
	}

	repo := event.GetRepo()
	commitSHA := event.GetHeadCommit().GetSHA()

	// installationID := githubapp.GetInstallationIDFromEvent(&event)
	// _, logger := githubapp.PrepareRepoContext(ctx, installationID, repo)

	repoName := repo.GetName()
	branchName := event.GetRef()
	//fullName := repo.GetFullName()
	owner := repo.GetOwner().GetLogin()
	org := repo.GetOrganization()
	if org == "" {
		org = owner
	}

	// TEST FILTER
	if org != "brondum" {
		fmt.Println(org)
		return nil
	}

	// create client
	client, err := gh.GetSingleOrganizationClient(h.ClientCreator, org)
	if err != nil {
		return err
	}

	// get content
	content, _, err := gh.GetRepoContent(ctx, client, owner, repoName, ".github/dependabot-circleci.yml", branchName)
	if err != nil {
		return nil // we dont care do we ?
	}

	checkName := "some"

	check, _, err := client.Checks.CreateCheckRun(ctx, owner, repoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: commitSHA,
		Status:  github.String("in_progress"),
	})

	// unmarshal
	var config config.RepoConfig
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		_, _, err := client.Checks.UpdateCheckRun(ctx, owner, repoName, check.GetID(), github.UpdateCheckRunOptions{
			Name:        checkName,
			Status:      github.String("completed"),
			Conclusion:  github.String("failure"),
			CompletedAt: &github.Timestamp{time.Now()},
			Output:      &github.CheckRunOutput{Title: github.String("FAILURE")}})
		if err != nil {
			return err
		}

		return nil
	}

	_, _, err = client.Checks.UpdateCheckRun(ctx, owner, repoName, check.GetID(), github.UpdateCheckRunOptions{
		Name:        checkName,
		HeadSHA:     github.String(commitSHA),
		Status:      github.String("completed"),
		Conclusion:  github.String("success"),
		CompletedAt: &github.Timestamp{time.Now()},
		Output: &github.CheckRunOutput{
			Title: github.String("GREAT SUCCESS"),
		}})
	if err != nil {
		return err
	}

	return nil
}
