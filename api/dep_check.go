package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/rs/zerolog/log"
)

var wg sync.WaitGroup

func dependencyHandler(w http.ResponseWriter, r *http.Request) {

	// dummy auth check, to decrease chance of ddos
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		http.Error(w, "please provide a valid bearer token", http.StatusUnauthorized)
		return
	}

	reqToken = strings.TrimSpace(splitToken[1])

	if reqToken != config.AppConfig.HTTP.Token {
		http.Error(w, "please provide a valid bearer token", http.StatusUnauthorized)
		return
	}

	// create clients
	clients, err := gh.GetOrganizationClients(config.AppConfig.Github)
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	// send stats to dd
	go datadog.Gauge("organizations", float64(len(clients)), nil)

	for _, client := range clients {
		wg.Add(1)
		client := client
		go func() {
			defer wg.Done()
			dependabot.Start(context.Background(), client)
		}()
	}
	wg.Wait()
	fmt.Fprintln(w, "Yaaay all done, please check github for pull requests!")
}
