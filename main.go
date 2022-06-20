package main

import (
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/BESTSELLER/dependabot-circleci/api"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/BESTSELLER/dependabot-circleci/logger"
	"github.com/BESTSELLER/go-vault/gcpss"

	"flag"
)

func init() {
	err := config.LoadEnvConfig()
	logger.Init()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}

	var secret []byte

	if config.EnvVars.Config == "" {
		vaultAddr := os.Getenv("VAULT_ADDR")
		if vaultAddr == "" {
			log.Fatal().Msg("VAULT_ADDR must be set")
		}
		vaultSecret := os.Getenv("VAULT_SECRET")
		if vaultSecret == "" {
			log.Fatal().Msg("VAULT_SECRET must be set")
		}
		vaultRole := os.Getenv("VAULT_ROLE")
		if vaultRole == "" {
			log.Fatal().Msg("VAULT_ROLE must be set")
		}

		secretData, err := gcpss.FetchVaultSecret(vaultAddr, vaultSecret, vaultRole)
		if err != nil {
			log.Fatal().Err(err)
		}
		secret = []byte(secretData)
	} else {
		bytes, err := ioutil.ReadFile(config.EnvVars.Config)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to read file %s", config.EnvVars.Config)
		}

		secret = bytes
	}

	err = config.ReadConfig([]byte(secret))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config:")
	}

}

func main() {

	standaloneFlag := flag.Bool("standalone", true, "Will run in standalone, which means that it will start all services.")
	webhookFlag := flag.Bool("webhook", false, "Will start the webhook server.")
	somethingFlag := flag.Bool("something", false, "Will start the something server.")
	controllerFlag := flag.Bool("controller", false, "Will start the controller.")

	flag.Parse()

	if *standaloneFlag && (*webhookFlag || *somethingFlag) {
		log.Fatal().Msg("-standalone is not allowed with any other flags.")
	}

	api.SetupRouter(*standaloneFlag, *webhookFlag, *somethingFlag, *controllerFlag)
}
