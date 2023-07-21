package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/rs/zerolog/log"
)

func dependencyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// extract repo details
	var workerPayload WorkerPayload
	err := json.NewDecoder(r.Body).Decode(&workerPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create client
	cc, err := gh.CreateGHClient(config.AppConfig.Github)
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client (gh.CreateGHClient)")
	}

	client, err := gh.GetSingleOrganizationClient(cc, workerPayload.Org)
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client (gh.GetSingleOrganizationClient)")
	}

	fmt.Fprintln(w, "Depedency check has started, please check github for incomming pull requests!")

	// do our magic
	dependabot.Start(context.Background(), client, workerPayload.Org, workerPayload.Repos)

	// send stats to DD
	defer datadog.TimeTrackAndGauge("dependency_check_duration", []string{fmt.Sprintf("organization:%s", workerPayload.Org)}, start)

	log.Debug().Msgf("Dependency check has completed for organization: %s", workerPayload.Org)
}
