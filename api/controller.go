package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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

var wg sync.WaitGroup

func controllerHandler(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msg("controllerHandler called")

	orgs, err := pullRepos()
	if err != nil {
		log.Error().Err(err).Msgf("pull repos from the db failed: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	log.Debug().Msgf("Found %d organizations", len(orgs))

	log.Debug().Msg("Sending metric to datadog")
	// send stats to dd
	go datadog.IncrementCount("organizations", int64(len(orgs)), nil)

	log.Debug().Msg("Triggering workers")
	// should be in parallel
	for organization, repositories := range orgs {
		wg.Add(1)
		go func(org string, repos []db.RepoData) {
			defer wg.Done()
			var triggeredRepos []string

			go datadog.IncrementCount("enabled_repos", int64(len(repos)), []string{fmt.Sprintf("organization:%s", org)})

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

			resp, err := PostJSON(fmt.Sprintf("%s/start", config.EnvVars.WorkerURL), payloadBytes)
			if err != nil {
				if resp != nil && resp.StatusCode == http.StatusNotAcceptable {
					log.Warn().Err(err).Msgf("dependabot-circleci not installed on org %s ", org)
					for _, repo := range repos {
						err := db.DeleteRepo(repo, context.Background())
						if err != nil {
							log.Error().Err(err).Msgf("error deleting repo %s", repo.Repo)
						}
					}
				} else {
					log.Error().Err(err).Msgf("error triggering worker for org %s", org)
				}
			} else {
				log.Debug().Msgf("Dependency check has started for org: %s", org)
			}
		}(organization, repositories)
	}

	wg.Wait()
	log.Debug().Msg("All workers finished")
}
func shouldRun(schedule string) bool {
	// check if an update should be run
	t := time.Now()
	schedule = strings.ToLower(schedule)
	switch schedule {
	case "daily", "":
		return true
	case "weekly":
		return t.Weekday() == 1
	case "monthly":
		return t.Day() == 1
	default:
		return false
	}
}

// PostJSON posts the structs as json to the specified url
func PostJSON(url string, payload []byte) (*http.Response, error) {
	var myClient *http.Client
	clientWithAuth, err := idtoken.NewClient(context.Background(), url)
	if err != nil {
		log.Warn().Err(err).Msg("Issues getting token from GCP metadata server, trying to continue without auth.")
		myClient = &http.Client{}
		// return fmt.Errorf("idtoken.NewClient: %v", err)
	} else {
		myClient = httptrace.WrapClient(clientWithAuth)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.HTTP.Token))
	r, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request: %s", err)
	}
	defer r.Body.Close()

	// check response code
	if r.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyString := string(bodyBytes)
		return r, fmt.Errorf("request failed, expected status: %d got: %d, error message: %s", http.StatusOK, r.StatusCode, bodyString)
	}
	return r, nil
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
