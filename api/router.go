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

// SetupRouter initializes the API routes
func SetupRouter(webhookEnabled bool, workerEnabled bool, controllerEnabled bool) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	server, err := baseapp.NewServer(
		config.AppConfig.Server,
		baseapp.DefaultParams(logger, fmt.Sprintf("%s.", appName))...,
	)
	if err != nil {
		logger.Panic().Err(err)
	}

	cc, err := githubapp.NewDefaultCachingClientCreator(
		config.AppConfig.Github,
		githubapp.WithClientUserAgent(fmt.Sprintf("%s/%s", appName, config.Version)),
		githubapp.WithClientTimeout(10*time.Second),
		githubapp.WithClientCaching(false, func() httpcache.Cache { return httpcache.NewMemoryCache() }),
		githubapp.WithClientMiddleware(
			githubapp.ClientMetrics(server.Registry()),
		),
	)
	if err != nil {
		logger.Panic().Err(err)
	}

	if webhookEnabled {
		webhookHandler := githubapp.NewEventDispatcher([]githubapp.EventHandler{&ConfigCheckHandler{ClientCreator: cc}}, config.AppConfig.Github.App.WebhookSecret, githubapp.WithScheduler(
			githubapp.AsyncScheduler(),
		))
		server.Mux().Handle(pat.Post("/"), webhookHandler)
	}

	if workerEnabled {
		server.Mux().HandleFunc(pat.Post("/start"), dependencyHandler)
	}

	if controllerEnabled {
		server.Mux().HandleFunc(pat.Post("/start_controller"), controllerHandler)
	}

	// Start is blocking
	err = server.Start()
	if err != nil {
		logger.Panic().Err(err)
	}
}
