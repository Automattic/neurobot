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
	"os"
	"strconv"
	"time"

	"github.com/apex/log"
)

var envFile = flag.String("env", "./.env", ".env file")

func main() {
	logger := log.Log
	flag.Parse()

	config, err := configuration.New(*envFile)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"envPath": *envFile,
		}).Fatal("Failed to load .env file")
	}

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	workflowsDefTOMLFile := os.Getenv("WORKFLOWS_DEF_TOML_FILE")

	logger.WithField("path", *envFile).Info("Loaded environment variables from .env")
	logger.Infof("Enabling debug? %t", config.Debug)
	logger.Infof("Using database file: %s", config.DatabasePath)

	databaseSession, err := database.MakeDatabaseSession()
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
	seeds.Bots(botRepository)

	// import TOML
	err = toml.Import(workflowsDefTOMLFile, workflowRepository, workflowStepsRepository)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"path": workflowsDefTOMLFile,
		}).Fatal("Failed to import TOML workflows")
	}

	botRegistry, err := makeBotRegistry(config.HomeserverName, botRepository)
	if err != nil {
		logger.WithError(err).Fatal("Failed to make bot registry")
	}

	// set default port for running webhook listener server
	webhookListenerPort, err := strconv.Atoi(os.Getenv("WEBHOOK_LISTENER_PORT"))
	if err != nil {
		webhookListenerPort = 8080
	}
	webhookListenerServer := http.NewServer(webhookListenerPort)

	// resolve .well-known to find our server URL to connect
	start := time.Now()
	serverURL := matrix.DiscoverServerURL(config.HomeserverName)
	logger.WithFields(log.Fields{
		"serverName": config.HomeserverName,
		"serverURL":  serverURL,
	}).WithDuration(time.Since(start)).Info("Discovered client API")

	e := engine.NewEngine(botRegistry, workflowStepsRepository)

	app := application.NewApp(e, botRegistry, workflowRepository, webhookListenerServer)
	if err := app.Run(); err != nil {
		logger.WithError(err).Fatal("Failed to run application")
	}

	logger.WithFields(log.Fields{
		"port": webhookListenerPort,
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
