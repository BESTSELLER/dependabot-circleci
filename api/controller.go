package api

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/bigquery"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

type bqdata struct {
	RepoName string
	Owner    string
	Org      string
}

func controllerHandler(w http.ResponseWriter, r *http.Request) {
	orgs, err := pullRepos()
	if err != nil {
		log.Err(err).Msgf("pull repos from big query failed: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
	}

	// send stats to dd
	go datadog.Gauge("organizations", float64(len(orgs)), nil)

	// should be in parralel
	for org, repos := range orgs {
		//start another container i guess
		fmt.Println(org)
		fmt.Println(repos)
	}

}

func pullRepos() (repos map[string][]string, err error) {
	ctx := context.Background()
	repos = make(map[string][]string)

	// Sets your Google Cloud Platform project ID.
	projectID := "dependabot-pub-prod-586e"

	// Creates a client.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal().Err(err).Msgf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	// pull everything, because we dare!
	q := client.Query("SELECT * FROM `dependabot_circleci.repos` ")

	it, err := q.Read(ctx)
	for {
		var row bqdata
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Err(err).Msgf("BQ fuckup: %s", err)
			return nil, err
		}

		repos[row.Org] = append(repos[row.Org], row.RepoName)
	}
	return repos, err
}
