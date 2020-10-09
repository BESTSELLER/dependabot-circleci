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

// create wait group
var wg sync.WaitGroup

func main() {
	config, err := config.ReadConfig("./config.yml")
	if err != nil {
		panic(err)
	}

	// create client
	client, err := gh.GetOrganizationClient(ctx, config.Github, "brondum")
	if err != nil {
		panic(err)
	}

	// get repos
	repos, err := gh.GetRepos(ctx, client, "brondum")
	if err != nil {
		panic(err)
	}

	// do magic
	for _, repository := range repos {
		wg.Add(1)
		go func(repo *github.Repository) {
			defer wg.Done()

			// get content of configfile
			content, SHA, err := gh.GetRepoContent(ctx, client, repo)
			if err != nil {
				return
			}

			// unmarshal
			var cciconfig yaml.Node
			err = yaml.Unmarshal(content, &cciconfig)
			if err != nil {
				panic(err)
			}

			// determine repo details
			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			baseBranch := repo.GetDefaultBranch()

			// check for updates
			updates := circleci.GetUpdates(&cciconfig)
			for old, update := range updates {
				newYaml := circleci.ReplaceVersion(update, old, string(content))

				// commit vars
				oldVersion := strings.Split(old, "@")
				newVersion := strings.Split(update.Value, "@")

				commitMessage := github.String(fmt.Sprintf("Bump @%s from %s to %s", oldVersion[0], oldVersion[1], newVersion[1]))
				commitBranch := github.String(fmt.Sprintf("dependabot-circleci/orb/%s", update.Value))

				// err := check and create branch
				err := gh.CreateBranch(ctx, client, repoOwner, repoName, baseBranch, commitBranch)
				if err != nil {
					log.Printf("could not create branch: %v", err)
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
				}

			}

		}(repository)
	}

	wg.Wait()
}
