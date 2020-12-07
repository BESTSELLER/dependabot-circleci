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

// ConfigCheckHandler handles all comments on issues
type ConfigCheckHandler struct {
	githubapp.ClientCreator
}

// Handles return list of events to listens to
func (h *ConfigCheckHandler) Handles() []string {
	return []string{"push"}
}

// Handle has ALL the logic! ;)
func (h *ConfigCheckHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse push event")
	}

	repo := event.GetRepo()
	commitSHA := event.GetAfter()
	repoName := repo.GetName()
	branchName := event.GetRef()
	owner := repo.GetOwner().GetLogin()
	org := repo.GetOrganization()
	if org == "" {
		org = owner
	}

	// TEST FILTER
	if org != "brondum" {
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

	checkName := "Check config"

	check, _, err := client.Checks.CreateCheckRun(ctx, owner, repoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: commitSHA,
		Status:  github.String("in_progress"),
	})

	// unmarshal
	var config config.RepoConfig
	err = yaml.UnmarshalStrict(content, &config)
	fmt.Println(config)
	if err != nil {
		_, _, err := client.Checks.UpdateCheckRun(ctx, owner, repoName, check.GetID(), github.UpdateCheckRunOptions{
			Name:        checkName,
			Status:      github.String("completed"),
			Conclusion:  github.String("failure"),
			CompletedAt: &github.Timestamp{time.Now()},
			Output: &github.CheckRunOutput{
				Title:   github.String("Failure"),
				Summary: github.String("The config is invalid: " + err.Error()),
			}})
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
			Title:   github.String("Success"),
			Summary: github.String("Congratulations, the config is valid"),
		}})
	if err != nil {
		return err
	}

	return nil
}
