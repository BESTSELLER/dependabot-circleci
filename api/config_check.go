package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/db"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type GithubInfo struct {
	RepoName string
	Owner    string
	Org      string
}

var Githubinfo GithubInfo

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
	start := time.Now()
	var event github.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return errors.Wrap(err, "failed to parse push event")
	}
	repo := event.GetRepo()
	commitSHA := event.GetAfter()
	Githubinfo.RepoName = repo.GetName()
	Githubinfo.Owner = repo.GetOwner().GetLogin()
	Githubinfo.Org = repo.GetOrganization()
	if Githubinfo.Org == "" {
		Githubinfo.Org = Githubinfo.Owner
	}

	// create client
	client, err := gh.GetSingleOrganizationClient(h.ClientCreator, Githubinfo.Org)
	if err != nil {
		log.Error().Err(err).Msgf("Inserting into bigquery table, had the following error: %s", err)
		return err
	}

	// get content
	content, _, err := gh.GetRepoContent(ctx, client, Githubinfo.Owner, Githubinfo.RepoName, ".github/dependabot-circleci.yml", commitSHA)
	if err != nil {
		log.Debug().Err(err).Msg("could not read content of repository")
		return nil // we dont care
	}

	checkName := "Check config"

	check, _, err := client.Checks.CreateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: commitSHA,
		Status:  github.String("in_progress"),
	})
	if err != nil {
		log.Error().Err(err).Msg("Error creating Github check")
		return err
	}

	// unmarshal
	var config config.RepoConfig
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		_, _, err := client.Checks.UpdateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, check.GetID(), github.UpdateCheckRunOptions{
			Name:        checkName,
			Status:      github.String("completed"),
			Conclusion:  github.String("failure"),
			CompletedAt: &github.Timestamp{Time: time.Now()},
			Output: &github.CheckRunOutput{
				Title:   github.String("Failure"),
				Summary: github.String("The configuration is invalid: " + err.Error()),
				Text:    github.String("Please refer to the [documentation](https://github.com/BESTSELLER/dependabot-circleci#getting-started) to setup a correct config file"),
			}})
		if err != nil {
			log.Error().Err(err).Msg("Error updating Github check with failed status")
			return err
		}

		return nil
	}

	// check config
	err = config.IsValid()
	if err != nil {
		_, _, err := client.Checks.UpdateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, check.GetID(), github.UpdateCheckRunOptions{
			Name:        checkName,
			Status:      github.String("completed"),
			Conclusion:  github.String("failure"),
			CompletedAt: &github.Timestamp{Time: time.Now()},
			Output: &github.CheckRunOutput{
				Title:   github.String("Failure"),
				Summary: github.String("The configuration is invalid: " + err.Error()),
				Text:    github.String("Please refer to the [documentation](https://github.com/BESTSELLER/dependabot-circleci#getting-started) to setup a correct config file"),
			}})
		if err != nil {
			log.Error().Err(err).Msg("Error updating Github check with failed status")
			return err
		}
	}

	// update github check with success state
	_, _, err = client.Checks.UpdateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, check.GetID(), github.UpdateCheckRunOptions{
		Name:        checkName,
		Status:      github.String("completed"),
		Conclusion:  github.String("success"),
		CompletedAt: &github.Timestamp{Time: time.Now()},
		Output: &github.CheckRunOutput{
			Title:   github.String("Success"),
			Summary: github.String("Congratulations, the configuration is valid"),
		}})
	if err != nil {
		log.Error().Err(err).Msg("Error updating Github check with success status")
		return err
	}

	defer datadog.TimeTrackAndHistogram("config_check_duration", []string{fmt.Sprintf("organization: %s", Githubinfo.Owner), fmt.Sprintf("repo: %s", Githubinfo.RepoName)}, start)

	return db.UpdateRepo(db.RepoData{
		ID:       repo.GetID(),
		Repo:     Githubinfo.RepoName,
		Owner:    Githubinfo.Owner,
		Schedule: config.Schedule,
	}, ctx)
}
