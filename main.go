package main

import (
	"context"
	"log"
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

			var cciconfig yaml.Node
			err = yaml.Unmarshal(content, &cciconfig)
			if err != nil {
				panic(err)
			}

			repoOwner := repo.GetOwner().GetLogin()
			repoName := repo.GetName()
			baseBranch := repo.GetDefaultBranch()

			// check for updates
			updates := circleci.GetUpdates(&cciconfig)
			for old, update := range updates {
				newYaml := circleci.ReplaceVersion(update, old, string(content))

				commitMessage := github.String("this is a test")
				commitBranch := github.String("test")

				var baseRef *github.Reference
				if baseRef, _, err = client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+baseBranch); err != nil {
					panic(err)
				}

				newRef := &github.Reference{Ref: github.String("refs/heads/" + *commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
				_, _, err := client.Git.CreateRef(ctx, repoOwner, repoName, newRef)
				if err != nil {
					panic(err)
				}

				_, _, err = client.Repositories.UpdateFile(ctx, repo.GetOwner().GetLogin(), repo.GetName(), ".circleci/config.yml",
					&github.RepositoryContentFileOptions{
						Message: commitMessage,
						Content: []byte(newYaml),
						Branch:  commitBranch,
						SHA:     SHA,
					})
				if err != nil {
					log.Printf("could not update file: %v", err)
				}

				_, _, err = client.PullRequests.Create(ctx, repoOwner, repoName, &github.NewPullRequest{
					Title:               github.String("TEST!"),
					Head:                commitBranch,
					Base:                github.String(baseBranch),
					Body:                commitMessage,
					MaintainerCanModify: github.Bool(true),
				})
				if err != nil {
					log.Printf("could not create pr: %v", err)
				}

			}

			// patch and pull request

		}(repository)
	}

	wg.Wait()
}
