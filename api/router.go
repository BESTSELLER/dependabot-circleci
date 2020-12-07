package api

import (
	"fmt"
	"os"
	"time"

	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/gregjones/httpcache"
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog"
	"goji.io/pat"
)

var appName = "dependabot-circleci"
var version = "1.0.0"

// SetupRouter initializes the API routes
func SetupRouter() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	appConfig, err := config.ReadConfig(config.EnvVars.Config)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read github app config")
	}

	server, err := baseapp.NewServer(
		appConfig.Server,
		baseapp.DefaultParams(logger, fmt.Sprintf("%s.", appName))...,
	)
	if err != nil {
		logger.Panic().Err(err)
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		appConfig.Github,
		githubapp.WithClientUserAgent(fmt.Sprintf("%s/%s", appName, version)),
		githubapp.WithClientTimeout(10*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(server.Registry()),
		),
	)
	if err != nil {
		logger.Panic().Err(err)
	}

	webhookHandler := githubapp.NewEventDispatcher([]githubapp.EventHandler{&ConfigCheckHandler{ClientCreator: cc}}, appConfig.Github.App.WebhookSecret, githubapp.WithScheduler(
		githubapp.AsyncScheduler(),
	))
	server.Mux().Handle(pat.Post("/"), webhookHandler)

	// Start is blocking
	err = server.Start()
	if err != nil {
		logger.Panic().Err(err)
	}
}
