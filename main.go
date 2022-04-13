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

	databaseSession, err := database.MakeDatabaseSession(config.DatabasePath)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"path": config.DatabasePath,
		}).Fatal("Failed to connect to database")
	}
	defer databaseSession.Close()
	err = database.Migrate(databaseSession)
	if err != nil {
		logger.WithError(err).Fatal("Failed to migrate database")
	}

	botRepository := botApp.NewRepository(databaseSession)
	workflowRepository := workflow.NewRepository(databaseSession)
	workflowStepsRepository := workflowstep.NewRepository(databaseSession)

	// Seed database.
	seeds.Bots(botRepository, config)

	// import TOML
	err = toml.Import(config.WorkflowsTOMLPath, workflowRepository, workflowStepsRepository)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"path": config.WorkflowsTOMLPath,
		}).Fatal("Failed to import TOML workflows")
	}

	botRegistry, err := makeBotRegistry(config.HomeserverName, botRepository)
	if err != nil {
		logger.WithError(err).Fatal("Failed to make bot registry")
	}

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

func makeBotRegistry(homeserverURL string, botRepository b.Repository) (registry botApp.Registry, err error) {
	bots, err := botRepository.FindActive()
	if err != nil {
		return
	}

	registry = botApp.NewRegistry(homeserverURL)

	for _, bot := range bots {
		var client matrix.Client
		client, err = matrix.NewMautrixClient(homeserverURL, true)
		if err != nil {
			return
		}

		err = registry.Append(bot, client)
		if err != nil {
			return
		}
	}

	return
}
