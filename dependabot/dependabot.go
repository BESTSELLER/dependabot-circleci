package dependabot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/BESTSELLER/dependabot-circleci/circleci"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v38/github"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var wg sync.WaitGroup

// Start will run through all repos it has access to and check for updates and make pull requests if needed.
func Start(ctx context.Context, client *github.Client) {
	// get repos
	repos, err := gh.GetRepos(ctx, client, 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to repos")
	}

	// Loop through all repos
	for _, repository := range repos {
		wg.Add(1)
		go checkRepo(ctx, client, repository)
	}
	wg.Wait()
}

func checkRepo(ctx context.Context, client *github.Client, repo *github.Repository) {
	defer wg.Done()

	if repo.GetArchived() != false {
		return
	}

	repoConfig := getRepoConfig(ctx, client, repo)
	if repoConfig == nil {
		return
	}

	// determine repo details
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	repoDefaultBranch := repo.GetDefaultBranch()

	targetBranch := getTargetBranch(ctx, client, repoOwner, repoName, repoDefaultBranch, repoConfig)
	if targetBranch == "" {
		return
	}

	go func() {
		datadog.IncrementCount("analysed_repos", repoOwner)
	}()

	// get content of circleci config file
	content, SHA, err := gh.GetRepoContent(ctx, client, repoOwner, repoName, ".circleci/config.yml", targetBranch)
	if err != nil {
		return
	}

	// unmarshal
	var cciconfig yaml.Node
	err = yaml.Unmarshal(content, &cciconfig)
	if err != nil {
		log.Error().Err(err).Msgf("could not unmarshal yaml in %s", repoName)
		return
	}

	// check for updates
	orbUpdates, dockerUpdates := circleci.GetUpdates(&cciconfig)
	for old, update := range orbUpdates {
		wg.Add(1)
		go handleUpdate(ctx, client, update, "orb", old, content, repoOwner, repoName, targetBranch, SHA, repoConfig)
	}
	for old, update := range dockerUpdates {
		wg.Add(1)
		go handleUpdate(ctx, client, update, "docker", old, content, repoOwner, repoName, targetBranch, SHA, repoConfig)
	}
}

func getRepoConfig(ctx context.Context, client *github.Client, repo *github.Repository) *config.RepoConfig {
	// check if a bot config exists
	repoConfigContent, _, err := gh.GetRepoContent(ctx, client, repo.GetOwner().GetLogin(), repo.GetName(), ".github/dependabot-circleci.yml", "")
	if err != nil {
		log.Debug().Err(err).Msgf("could not load dependabot-circleci.yml in repo: %s", repo.GetName())
		return nil
	}

	repoConfig, err := config.ReadRepoConfig(repoConfigContent)
	if err != nil {
		log.Error().Err(err).Msgf("could not read repo config in %s", repo.GetName())
		return nil
	}

	return repoConfig
}
func getTargetBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, defaultBranch string, repoConfig *config.RepoConfig) string {
	targetBranch := defaultBranch
	if repoConfig.TargetBranch != "" {
		_, _, err := client.Repositories.GetBranch(ctx, repoOwner, repoName, repoConfig.TargetBranch, true)
		if err != nil {
			return ""
		}
		targetBranch = repoConfig.TargetBranch
	}
	return targetBranch
}

func handleUpdate(ctx context.Context, client *github.Client, update *yaml.Node, updateType string, old string, content []byte, repoOwner string, repoName string, targetBranch string, SHA *string, repoConfig *config.RepoConfig) {
	defer wg.Done()

	log.Debug().Msgf("repo: %s, old: %s, update: %s", repoName, old, update.Value)
	newYaml := circleci.ReplaceVersion(update, old, string(content))

	// commit vars
	oldVersion, newVersion := []string{}, []string{}
	if updateType == "orb" {
		oldVersion = strings.Split(old, "@")
		newVersion = strings.Split(update.Value, "@")
	} else {
		oldVersion = strings.Split(old, ":")
		newVersion = strings.Split(update.Value, ":")
	}

	if updateType == "orb" && len(newVersion) == 1 {
		return
	}

	commitMessage := fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1])
	commitBranch := fmt.Sprintf("dependabot-circleci/%s/%s", updateType, strings.ReplaceAll(update.Value, ":", "@"))

	// err := check and create branch
	exists, oldPR, err := gh.CheckPR(ctx, client, repoOwner, repoName, targetBranch, commitBranch, commitMessage, oldVersion[0])
	if err != nil {
		log.Error().Err(err).Msgf("could not get old branch in %s", repoName)
		return
	}
	if exists {
		return
	}
	err = gh.CreateBranch(ctx, client, repoOwner, repoName, targetBranch, github.String(commitBranch))
	if err != nil {
		log.Error().Err(err).Msgf("could not create branch %s in %s", commitBranch, repoName)
		return
	}

	// commit file
	err = gh.UpdateFile(ctx, client, repoOwner, repoName, &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: []byte(newYaml),
		Branch:  github.String(commitBranch),
		SHA:     SHA,
	})
	if err != nil {
		log.Error().Err(err).Msgf("could not update file in %s", repoName)
		return
	}

	// create pull req
	newPR, err := gh.CreatePR(ctx, client, repoOwner, repoName, repoConfig.Reviewers, repoConfig.Assignees, repoConfig.Labels, &github.NewPullRequest{
		Title:               github.String(commitMessage),
		Head:                github.String(commitBranch),
		Base:                github.String(targetBranch),
		Body:                github.String(commitMessage),
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		log.Info().Err(err).Msgf("could not create pr in %s", repoName)
		return
	}

	go func() {
		datadog.IncrementCount("pull_requests", repoOwner)
	}()

	if oldPR != nil {
		err := gh.CleanUpOldBranch(ctx, client, repoOwner, repoName, oldPR, newPR.GetNumber())
		if err != nil {
			log.Error().Err(err).Msgf("could not cleanup old pr and branch in %s", repoName)
			return
		}

		go func() {
			datadog.IncrementCount("superseeded_updates", repoOwner)
		}()
	}
}
