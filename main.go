package main

import (
	"fmt"
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
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}
	logger.Init()
	fmt.Printf("Logging level: %d\n", *config.EnvVars.LogLevel)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to read env config")
	}

	var appsecret []byte
	var dbsecret []byte

	if config.EnvVars.Config == "" {
		fmt.Println("No config file specified, fetching secrets from vault")
		vaultAddr := os.Getenv("VAULT_ADDR")
		if vaultAddr == "" {
			log.Fatal().Msg("VAULT_ADDR must be set")
		}
		vaultRole := os.Getenv("VAULT_ROLE")
		if vaultRole == "" {
			log.Fatal().Msg("VAULT_ROLE must be set")
		}

		appSecret := os.Getenv("APP_SECRET")
		if appSecret == "" {
			log.Fatal().Msg("APP_SECRET must be set")
		}

		dbSecret := os.Getenv("DB_SECRET")
		if dbSecret == "" {
			log.Fatal().Msg("DB_SECRET must be set")
		}

		fmt.Printf("VAULT_ADDR: %s\nVAULT_ROLE: %s\nAPP_SECRET: %s\nDB_SECRET: %s\n", vaultAddr, vaultRole, appSecret, dbSecret)

		// fetch app secrets
		secretData, err := gcpss.FetchVaultSecret(vaultAddr, appSecret, vaultRole)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to fetch secrets from vault. error %v", err)
		}
		appsecret = []byte(secretData)

		// fetch db secrets
		secretData, err = gcpss.FetchVaultSecret(vaultAddr, dbSecret, vaultRole)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to fetch secrets from vault. error %v", err)
		}
		dbsecret = []byte(secretData)

	} else {
		fmt.Println("Using config file: ", config.EnvVars.Config)
		bytes, err := ioutil.ReadFile(config.EnvVars.Config)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to read file %s", config.EnvVars.Config)
		}
		appsecret = bytes

		fmt.Println("Using db config file: ", config.EnvVars.DBConfig)
		bytes, err = ioutil.ReadFile(config.EnvVars.DBConfig)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to read file %s", config.EnvVars.DBConfig)
		}
		dbsecret = bytes
	}

	err = config.ReadAppConfig([]byte(appsecret))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read github app config:")
	}

	err = config.ReadDBConfig([]byte(dbsecret))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read db config:")
	}

}

func main() {

	webhookFlag := flag.Bool("webhook", false, "Will start the webhook server.")
	workerFlag := flag.Bool("worker", false, "Will start the worker server.")
	controllerFlag := flag.Bool("controller", false, "Will start the controller.")

	flag.Parse()

	api.SetupRouter(*webhookFlag, *workerFlag, *controllerFlag)
}
