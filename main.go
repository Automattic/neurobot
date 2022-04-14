package main

import (
	"flag"
	application "neurobot/app"
	botApp "neurobot/app/bot"
	configuration "neurobot/app/config"
	"neurobot/app/engine"
	"neurobot/app/workflow"
	"neurobot/app/workflowstep"
	"neurobot/infrastructure/database"
	"neurobot/infrastructure/http"
	"neurobot/infrastructure/matrix"
	"neurobot/infrastructure/toml"
	b "neurobot/model/bot"
	"neurobot/resources/seeds"
	"strings"

	"github.com/apex/log"
)

var envFile = flag.String("env", "./.env", ".env file")

func main() {
	logger := log.Log
	flag.Parse()
	config := configuration.LoadFromEnvFile(*envFile)

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	databaseSession := database.MakeDatabaseSession(config.DatabasePath)
	defer databaseSession.Close()

	botRepository := botApp.NewRepository(databaseSession)
	workflowRepository := workflow.NewRepository(databaseSession)
	workflowStepsRepository := workflowstep.NewRepository(databaseSession)

	// Seed database.
	seeds.Bots(botRepository, config)

	// import TOML
	err := toml.Import(config.WorkflowsTOMLPath, workflowRepository, workflowStepsRepository)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"path": config.WorkflowsTOMLPath,
		}).Fatal("Failed to import TOML workflows")
	}

	botRegistry := makeBotRegistry(config.HomeserverName, botRepository)
	webhookListenerServer := http.NewServer(config.WebhookListenerPort)

	e := engine.NewEngine(botRegistry, workflowStepsRepository)

	app := application.NewApp(e, botRegistry, workflowRepository, webhookListenerServer)
	if err := app.Run(); err != nil {
		logger.WithError(err).Fatal("Failed to run application")
	}

	logger.WithFields(log.Fields{
		"port": config.WebhookListenerPort,
	}).Infof("Starting webhook listener")
	webhookListenerServer.Run() // blocking
}

func makeBotRegistry(homeserverName string, botRepository b.Repository) (registry botApp.Registry) {
	homeserverURL, err := matrix.DiscoverServerURL(homeserverName)
	if err != nil {
		log.WithError(err).Fatal("Failed to discover homeserver URL")
	}

	bots, err := botRepository.FindActive()
	if err != nil {
		log.WithError(err).Fatal("Failed to find active bots")
	}

	homeserverDomain := strings.Split(homeserverURL.Host, ":")[0]
	registry = botApp.NewRegistry(homeserverDomain)

	for _, bot := range bots {
		var client matrix.Client
		client, err = matrix.NewMautrixClient(homeserverURL, true)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"username": bot.Username,
			}).Fatal("Failed to login as bot")
		}

		err = registry.Append(bot, client)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"username": bot.Username,
			}).Fatal("Failed add bot to registry")
		}
	}

	return
}
