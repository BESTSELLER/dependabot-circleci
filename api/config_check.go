package api

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v41/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type GithubInfo struct {
	RepoName  string
	Owner string
	Org string
}
var Githubinfo GithubInfo

func update_bigquery() {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "dependabot-pub-prod-586e"

	// Creates a client.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal().Err(err).Msgf("bigquery.NewClient: %v", err)
	}
	defer client.Close()
	ins := client.Dataset("dependabot_circleci").Table("repos").Inserter()
	if err := ins.Put(ctx, Githubinfo); err != nil {
		log.Err(err).Msgf("Inserting into bigquery table, had the following error: %s", err)
	}
	log.Debug().Msg("All done bigquery table updated")
}


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
	Githubinfo.RepoName = repo.GetName()
	Githubinfo.Owner = repo.GetOwner().GetLogin()
	Githubinfo.Org = repo.GetOrganization()
	if Githubinfo.Org == "" {
		Githubinfo.Org = Githubinfo.Owner
	}

	// create client
	client, err := gh.GetSingleOrganizationClient(h.ClientCreator, Githubinfo.Org)
	if err != nil {
		return err
	}

	// get content
	content, _, err := gh.GetRepoContent(ctx, client, Githubinfo.Owner, Githubinfo.RepoName , ".github/dependabot-circleci.yml", commitSHA)
	if err != nil {
		return nil // we dont care
	}

	checkName := "Check config"

	check, _, err := client.Checks.CreateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: commitSHA,
		Status:  github.String("in_progress"),
	})

	// unmarshal
	var config config.RepoConfig
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		_, _, err := client.Checks.UpdateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, check.GetID(), github.UpdateCheckRunOptions{
			Name:        checkName,
			Status:      github.String("completed"),
			Conclusion:  github.String("failure"),
			CompletedAt: &github.Timestamp{time.Now()},
			Output: &github.CheckRunOutput{
				Title:   github.String("Failure"),
				Summary: github.String("The configuration is invalid: " + err.Error()),
				Text:    github.String("Please refer to the [documentation](https://github.com/BESTSELLER/dependabot-circleci#getting-started) to setup a correct config file"),
			}})
		if err != nil {
			return err
		}

		return nil
	}

	_, _, err = client.Checks.UpdateCheckRun(ctx, Githubinfo.Owner, Githubinfo.RepoName, check.GetID(), github.UpdateCheckRunOptions{
		Name:        checkName,
		Status:      github.String("completed"),
		Conclusion:  github.String("success"),
		CompletedAt: &github.Timestamp{time.Now()},
		Output: &github.CheckRunOutput{
			Title:   github.String("Success"),
			Summary: github.String("Congratulations, the configuration is valid"),
		}})
	if err != nil {
		return err
	}
	update_bigquery()
	return nil
}
