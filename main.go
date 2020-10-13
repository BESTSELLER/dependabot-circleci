package main

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

var ctx = context.Background()
var wg sync.WaitGroup
var org = "brondum"

func main() {
	appConfig, err := config.ReadConfig("./config.yml")
	if err != nil {
		panic(err)
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, appConfig.Github, org)
	if err != nil {
		panic(err)
	}

	// get repos
	repos, err := gh.GetRepos(ctx, client)
	if err != nil {
		panic(err)
	}

	// do magic
	for _, repository := range repos {
		wg.Add(1)
		go func(repo *github.Repository) {
			defer wg.Done()

			// check if a bot config exists
			repoConfigContent, _, err := gh.GetRepoContent(ctx, client, repo, ".github/dependabot-circleci.yml", "")
			if err != nil {
				return
			}

			repoConfig, err := config.ReadRepoConfig(repoConfigContent)
			if err != nil {
				log.Println(err)
				return
			}

			// determine repo details
			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()

			targetBranch := repo.GetDefaultBranch()
			if repoConfig.TargetBranch != "" {
				_, _, err := client.Repositories.GetBranch(ctx, repoOwner, repoName, repoConfig.TargetBranch)
				if err != nil {
					return
				}
				targetBranch = repoConfig.TargetBranch
			}

			// get content of circleci configfile
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
				fmt.Println(update)
				newYaml := circleci.ReplaceVersion(update, old, string(content))

				// commit vars
				oldVersion := strings.Split(old, "@")
				newVersion := strings.Split(update.Value, "@")
				commitMessage := github.String(fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1]))
				commitBranch := github.String(fmt.Sprintf("dependabot-circleci/orb/%s", update.Value))

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

		}(repository)
	}

	wg.Wait()
}
