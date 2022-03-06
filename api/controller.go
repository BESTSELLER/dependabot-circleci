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

func controllerHandler(w http.ResponseWriter, r *http.Request) {
	orgs, err := pullOrgs()
	if err != nil {
		log.Err(err).Msgf("pull orgs from big query failed: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
	}

	// send stats to dd
	go datadog.Gauge("organizations", float64(len(orgs)), nil)

	// should be in parralel
	for _, org := range orgs {
		//start another container i guess
		fmt.Println(org)
	}

}

func pullOrgs() (orgs []string, err error) {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "dependabot-pub-prod-586e"

	// Creates a client.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal().Err(err).Msgf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	// select distinct orgs
	q := client.Query(
		"SELECT org FROM `dependabot_circleci.repos` " +
			"GROUP BY org ")

	it, err := q.Read(ctx)
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Err(err).Msgf("BQ fuckup: %s", err)
			return nil, err
		}
		// have to range each row as they are a array in itself?!?
		for _, value := range row {
			orgs = append(orgs, fmt.Sprintf("%s", value))
		}
	}
	return orgs, err
}
