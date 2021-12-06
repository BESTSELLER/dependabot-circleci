package main

import (
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/api"
	"github.com/BESTSELLER/dependabot-circleci/logger"

	"github.com/BESTSELLER/dependabot-circleci/config"
)

var wg sync.WaitGroup

// TODO remeber to uncomment
// func init() {
// 	vaultAddr := os.Getenv("VAULT_ADDR")
// 	if vaultAddr == "" {
// 		log.Fatal().Msg("VAULT_ADDR must be set")
// 	}
// 	vaultSecret := os.Getenv("VAULT_SECRET")
// 	if vaultSecret == "" {
// 		log.Fatal().Msg("VAULT_SECRET must be set")
// 	}
// 	vaultRole := os.Getenv("VAULT_ROLE")
// 	if vaultRole == "" {
// 		log.Fatal().Msg("VAULT_ROLE must be set")
// 	}

// 	secret, err := gcpss.FetchVaultSecret(vaultAddr, vaultSecret, vaultRole)
// 	if err != nil {
// 		log.Fatal().Err(err)
// 	}

// 	err = os.Mkdir("/secrets", 0644)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("unable to create secrets dir")
// 	}

// 	data := []byte(secret)
// 	err = ioutil.WriteFile("/secrets/secrets", data, 0644)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("unable to write secrets")
// 	}

// }

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

	// start webhook
	api.SetupRouter()

}
