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

	webhookFlag := flag.Bool("webhook", false, "Will start the webhook server.")
	workerFlag := flag.Bool("worker", false, "Will start the worker server.")
	controllerFlag := flag.Bool("controller", false, "Will start the controller.")

	flag.Parse()

	api.SetupRouter(*webhookFlag, *workerFlag, *controllerFlag)
}
