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
	"github.com/BESTSELLER/dependabot-circleci/db"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/rs/zerolog/log"

	"google.golang.org/api/idtoken"
)

type WorkerPayload struct {
	Org   string
	Repos []string
}

func controllerHandler(w http.ResponseWriter, r *http.Request) {
	orgs, err := pullRepos()
	if err != nil {
		log.Error().Err(err).Msgf("pull repos from big query failed: %s", err)
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
				triggeredRepos = append(triggeredRepos, repo.Repo)
			}
		}
		payloadObj := WorkerPayload{Org: org, Repos: triggeredRepos}
		payloadBytes, err := json.Marshal(payloadObj)
		if err != nil {
			log.Error().Err(err).Msg("error marshaling payload")
			return
		}
		err = PostJSON(fmt.Sprintf("%s/start", config.EnvVars.WorkerURL), payloadBytes)
		if err != nil {
			log.Error().Err(err).Msgf("error triggering worker for org %s", org)
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

	clientWithAuth, err := idtoken.NewClient(context.Background(), url)
	if err != nil {
		return fmt.Errorf("idtoken.NewClient: %v", err)
	}

	var myClient = httptrace.WrapClient(clientWithAuth)

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

func pullRepos() (repos map[string][]db.RepoData, err error) {
	ctx := context.Background()
	repos = make(map[string][]db.RepoData)

	repoList, err := db.GetRepos(ctx)
	if err != nil {
		return nil, err
	}

	for _, repo := range repoList {
		repos[repo.Owner] = append(repos[repo.Owner], repo)
	}

	return repos, err
}
