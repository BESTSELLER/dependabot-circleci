package dependabot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

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
	repos, err := gh.GetRepos(ctx, client)
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
		newYaml := circleci.ReplaceVersion(update, old, string(content))

		// commit vars
		oldVersion := strings.Split(old, "@")
		newVersion := strings.Split(update.Value, "@")
		commitMessage := github.String(fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1]))
		commitBranch := github.String(fmt.Sprintf("dependabot-circleci/orb/%s", update.Value))

		fmt.Println(*commitMessage)

		// err := check and create branch
		exists, oldPR, err := gh.CheckPR(ctx, client, repoOwner, repoName, targetBranch, commitBranch, commitMessage, oldVersion[0])
		if err != nil {
			log.Printf("could not get old branch: %v", err)
			continue
		}
		if exists {
			continue
		}
		err = gh.CreateBranch(ctx, client, repoOwner, repoName, targetBranch, commitBranch)
		if err != nil {
			log.Printf("could not create branch: %v", err)
			continue
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
			continue
		}

		// create pull req
		newPR, err := gh.CreatePR(ctx, client, repoOwner, repoName, repoConfig.DefaultReviewers, repoConfig.DefaultAssignees, repoConfig.DefaultLabels, &github.NewPullRequest{
			Title:               commitMessage,
			Head:                commitBranch,
			Base:                github.String(targetBranch),
			Body:                commitMessage,
			MaintainerCanModify: github.Bool(true),
		})
		if err != nil {
			log.Printf("could not create pr: %v", err)
			continue
		}

		if oldPR != nil {
			err := gh.CleanUpOldBranch(ctx, client, repoOwner, repoName, oldPR, newPR.GetNumber())
			if err != nil {
				log.Printf("could not cleanup old pr and branch: %v", err)
				continue
			}
		}

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
