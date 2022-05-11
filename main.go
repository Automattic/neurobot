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
	"neurobot/model/command"
	"neurobot/resources/seeds"
	"strings"

	"github.com/apex/log"
	"github.com/upper/db/v4"
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

	commandsHandler := command.NewCommandsHandler()
	botRegistry := makeBotRegistry(config.ServerName, botRepository, databaseSession, commandsHandler)
	webhookListenerServer := http.NewServer(config.WebhookListenerPort)

	app := application.NewApp(
		engine.NewEngine(botRegistry, workflowStepsRepository),
		botRegistry,
		workflowRepository,
		webhookListenerServer,
		commandsHandler,
	)
	if err := app.Run(); err != nil {
		logger.WithError(err).Fatal("Failed to run application")
	}
}

func makeBotRegistry(serverName string, botRepository b.Repository, db db.Session, ch command.CommandsHandler) (registry botApp.Registry) {
	homeserverURL, err := matrix.DiscoverServerURL(serverName)
	if err != nil {
		log.WithError(err).Fatal("Failed to discover homeserver URL")
	}

	bots, err := botRepository.FindActive()
	if err != nil {
		log.WithError(err).Fatal("Failed to find active bots")
	}

	serverNameWithoutPort := strings.Split(serverName, ":")[0]
	registry = botApp.NewRegistry(serverNameWithoutPort, ch)

	for _, bot := range bots {
		storer := matrix.NewStorer(db, bot.ID)
		var client matrix.Client
		client, err = matrix.NewMautrixClient(homeserverURL, storer, true)
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
