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
	config, err := config.ReadConfig("./config.yml")
	if err != nil {
		panic(err)
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, config.Github, org)
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
			_, _, err := gh.GetRepoContent(ctx, client, repo, ".github/circleci.yml")
			if err != nil {
				return
			}

			// get content of circleci configfile
			content, SHA, err := gh.GetRepoContent(ctx, client, repo, ".circleci/config.yml")
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

			// determine repo details
			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			baseBranch := repo.GetDefaultBranch()

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
				exists, err := gh.CheckPR(ctx, client, repoOwner, repoName, baseBranch, commitBranch, commitMessage, oldVersion[0])
				if err != nil {
					log.Printf("could not create branch: %v", err)
					continue
				}
				if exists {
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
				err = gh.CreatePR(ctx, client, repoOwner, repoName, &github.NewPullRequest{
					Title:               commitMessage,
					Head:                commitBranch,
					Base:                github.String(baseBranch),
					Body:                commitMessage,
					MaintainerCanModify: github.Bool(true),
				})
				if err != nil {
					log.Printf("could not create pr: %v", err)
					continue
				}

			}

		}(repository)
	}

	wg.Wait()
}
