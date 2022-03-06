package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/gh"
	"github.com/rs/zerolog/log"
)

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

	//extract body and parse repos and org
	// var some struct {
	// 	org   string
	// 	repos []repo
	// }

	// create client
	cc, err := gh.CreateGHClient(config.AppConfig.Github)
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	client, err := gh.GetSingleOrganizationClient(cc, "org")
	if err != nil {
		http.Error(w, "failed to register organization client", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("failed to register organization client")
	}

	dependabot.Start(context.Background(), client)

	fmt.Fprintln(w, "Yaaay all done, please check github for pull requests!")
}
