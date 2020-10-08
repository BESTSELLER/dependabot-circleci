package main

import (
	"context"
	"fmt"
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
			content, err := gh.GetRepoContent(ctx, client, repo)
			if err != nil {
				return
			}

			var cciconfig yaml.Node
			err = yaml.Unmarshal(content, &cciconfig)
			if err != nil {
				panic(err)
			}

			// check for updates
			updates := circleci.GetUpdates(&cciconfig)
			circleci.ReplaceVersion(updates[0], string(content))
			fmt.Printf("%+v", updates[0])

			// patch and pull request

		}(repository)

		// by, err := yaml.Marshal(&some)
		// if err != nil {
		// 	panic(err)
		// }

		// err = ioutil.WriteFile("output.yml", by, 0644)
		// if err != nil {
		// 	panic(err)
		// }
	}

	wg.Wait()
}
