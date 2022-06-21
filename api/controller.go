package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"cloud.google.com/go/bigquery"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

type bqdata struct {
	RepoName string
	Owner    string
	Schedule string
}
type WorkerPayload struct {
	Org   string
	Repos []string
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
		var triggeredRepos []string

		go datadog.Gauge("enabled_repos", float64(len(repos)), []string{org})

		for _, repo := range repos {
			if shouldRun(repo.Schedule) {
				triggeredRepos = append(triggeredRepos, repo.RepoName)
			}
		}
		payloadObj := WorkerPayload{Org: org, Repos: triggeredRepos}
		payloadBytes, err := json.Marshal(payloadObj)
		if err != nil {
			log.Err(err).Msg("error marshaling payload")
			return
		}
		err = PostJSON(fmt.Sprintf("%s/start", config.EnvVars.WorkerURL), payloadBytes)
		if err != nil {
			log.Err(err).Msgf("error triggering worker for org %s", org)
		}
		// call webhook - trigger cci on org
		//start another container i guess
	}

}
func shouldRun(schedule string) bool {
	// check if an update should be run
	t := time.Now()
	schedule = strings.ToLower(schedule)
	if schedule == "monthly" {
		if t.Day() == 1 {
			return true
		}
		return false
	} else if schedule == "weekly" {
		if t.Weekday() == 1 {
			return true
		}
		return false

	} else if schedule == "daily" || schedule == "" {
		return true
	}
	return false
}

// PostJSON posts the structs as json to the specified url
func PostJSON(url string, payload []byte) error {

	var myClient = httptrace.WrapClient(&http.Client{Timeout: 30 * time.Second})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Unable to create request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.HTTP.Token))
	r, err := myClient.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err)
	}
	defer r.Body.Close()

	// check response code
	if r.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		bodyString := string(bodyBytes)
		return fmt.Errorf("Request failed, expected status: %d got: %d, error message: %s", http.StatusOK, r.StatusCode, bodyString)
	}
	return nil
}

func pullRepos() (repos map[string][]bqdata, err error) {
	ctx := context.Background()
	repos = make(map[string][]bqdata)

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

		repos[row.Owner] = append(repos[row.Owner], row)
	}
	return repos, err
}
