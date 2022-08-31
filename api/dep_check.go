package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/rs/zerolog/log"
)

func dependencyHandler(w http.ResponseWriter, r *http.Request) {

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
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	client, err := gh.GetSingleOrganizationClient(cc, workerPayload.Org)
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	// do our magic
	dependabot.Start(context.Background(), client, workerPayload.Repos)

	fmt.Fprintln(w, "Yaaay all done, please check github for pull requests!")
}
