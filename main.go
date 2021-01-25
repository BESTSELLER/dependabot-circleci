package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/api"
	"github.com/BESTSELLER/dependabot-circleci/datadog"
	"github.com/BESTSELLER/dependabot-circleci/dependabot"
	"github.com/BESTSELLER/dependabot-circleci/logger"
	"github.com/BESTSELLER/go-vault/gcpss"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/gh"
)

var wg sync.WaitGroup

func init() {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		fmt.Println("VAULT_ADDR must be set.")
	}
	vaultSecret := os.Getenv("VAULT_SECRET")
	if vaultSecret == "" {
		fmt.Println("VAULT_SECRET must be set.")
	}
	vaultRole := os.Getenv("VAULT_ROLE")
	if vaultRole == "" {
		fmt.Println("VAULT_ROLE must be set.")
	}

	secret, err := gcpss.FetchVaultSecret(vaultAddr, vaultSecret, vaultRole)
	if err != nil {
		fmt.Println(err)
	}

	err = os.Mkdir("/secrets", 0644)
	if err != nil {
		fmt.Println(err)
	}

	data := []byte(secret)
	err = ioutil.WriteFile("/secrets/secrets", data, 0644)
	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	err := config.LoadEnvConfig()
	logger.Init()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}

	err = config.ReadConfig(config.EnvVars.Config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config:")
	}

	//schedule checks
	scheduleTime := config.EnvVars.Schedule
	if scheduleTime == "" {
		scheduleTime = "22:00"
	}

	schedule := gocron.NewScheduler(time.UTC)
	_, err = schedule.Every(1).Day().At(scheduleTime).Do(runDependabot)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create schedule")
	}
	schedule.StartAsync()

	_, next := schedule.NextRun()
	log.Info().Msgf("Next scheduled dependency check is at: %s", next)

	// start webhook
	api.SetupRouter()

}

func runDependabot() {
	// create clients
	clients, err := gh.GetOrganizationClients(config.AppConfig.Github)
	if err != nil {
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
}
