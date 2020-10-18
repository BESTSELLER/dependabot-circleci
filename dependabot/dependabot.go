package dependabot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/BESTSELLER/dependabot-circleci/datadog"

	"github.com/BESTSELLER/dependabot-circleci/circleci"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/google/go-github/v32/github"
	"gopkg.in/yaml.v3"
)

var wg sync.WaitGroup

// Start will run through all repos it has access to and check for updates and make pull requests if needed.
func Start(ctx context.Context, client *github.Client) {
	// get repos
	repos, err := gh.GetRepos(ctx, client, 1)
	if err != nil {
		panic(err)
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
	content, SHA, err := gh.GetRepoContent(ctx, client, repo, ".circleci/config.yml", targetBranch)
	if err != nil {
		return
	}

	// unmarshal
	var cciconfig yaml.Node
	err = yaml.Unmarshal(content, &cciconfig)
	if err != nil {
		log.Printf("could not unmarshal yaml: %v", err)
		return
	}

	// check for updates
	updates := circleci.GetUpdates(&cciconfig)
	for old, update := range updates {
		wg.Add(1)
		go handleUpdate(ctx, client, update, old, content, repoOwner, repoName, targetBranch, SHA, repoConfig)
	}
}
func getRepoConfig(ctx context.Context, client *github.Client, repo *github.Repository) *config.RepoConfig {
	// check if a bot config exists
	repoConfigContent, _, err := gh.GetRepoContent(ctx, client, repo, ".github/dependabot-circleci.yml", "")
	if err != nil {
		return nil
	}

	repoConfig, err := config.ReadRepoConfig(repoConfigContent)
	if err != nil {
		log.Println(err)
		return nil
	}

	return repoConfig
}
func getTargetBranch(ctx context.Context, client *github.Client, repoOwner string, repoName string, defaultBranch string, repoConfig *config.RepoConfig) string {
	targetBranch := defaultBranch
	if repoConfig.TargetBranch != "" {
		_, _, err := client.Repositories.GetBranch(ctx, repoOwner, repoName, repoConfig.TargetBranch)
		if err != nil {
			return ""
		}
		targetBranch = repoConfig.TargetBranch
	}
	return targetBranch
}

func handleUpdate(ctx context.Context, client *github.Client, update *yaml.Node, old string, content []byte, repoOwner string, repoName string, targetBranch string, SHA *string, repoConfig *config.RepoConfig) {
	defer wg.Done()

	fmt.Printf("repo: %s, old: %s, update: %s\n", repoName, old, update.Value)
	newYaml := circleci.ReplaceVersion(update, old, string(content))

	// commit vars
	oldVersion := strings.Split(old, "@")
	newVersion := strings.Split(update.Value, "@")

	if len(newVersion) == 1 {
		return
	}

	commitMessage := github.String(fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1]))
	commitBranch := github.String(fmt.Sprintf("dependabot-circleci/orb/%s", update.Value))

	// err := check and create branch
	exists, oldPR, err := gh.CheckPR(ctx, client, repoOwner, repoName, targetBranch, commitBranch, commitMessage, oldVersion[0])
	if err != nil {
		log.Printf("could not get old branch: %v", err)
		return
	}
	if exists {
		return
	}
	err = gh.CreateBranch(ctx, client, repoOwner, repoName, targetBranch, commitBranch)
	if err != nil {
		log.Printf("could not create branch: %v", err)
		return
	}

	// commit file
	err = gh.UpdateFile(ctx, client, repoOwner, repoName, &github.RepositoryContentFileOptions{
		Message: commitMessage,
		Content: []byte(newYaml),
		Branch:  commitBranch,
		SHA:     SHA,
	})
	if err != nil {
		log.Printf("could not update file: %v", err)
		return
	}

	// create pull req
	newPR, err := gh.CreatePR(ctx, client, repoOwner, repoName, repoConfig.Reviewers, repoConfig.Assignees, repoConfig.Labels, &github.NewPullRequest{
		Title:               commitMessage,
		Head:                commitBranch,
		Base:                github.String(targetBranch),
		Body:                commitMessage,
		MaintainerCanModify: github.Bool(true),
	})
	if err != nil {
		log.Printf("could not create pr: %v", err)
		return
	}

	go func() {
		datadog.IncrementCount("pull_requests", repoOwner)
	}()

	if oldPR != nil {
		err := gh.CleanUpOldBranch(ctx, client, repoOwner, repoName, oldPR, newPR.GetNumber())
		if err != nil {
			log.Printf("could not cleanup old pr and branch: %v", err)
			return
		}

		go func() {
			datadog.IncrementCount("superseeded_updates", repoOwner)
		}()
	}
}
