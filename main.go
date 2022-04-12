package main

import (
	"flag"
	application "neurobot/app"
	botApp "neurobot/app/bot"
	"neurobot/app/workflow"
	"neurobot/app/workflowstep"
	"neurobot/engine"
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
	"github.com/joho/godotenv"
)

var envFile = flag.String("env", "./.env", ".env file")

func main() {
	logger := log.Log
	flag.Parse()

	err := godotenv.Load(*envFile)
	if err != nil {
		logger.WithError(err).Fatal("Error loading .env file")
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false // default
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	dbFile := os.Getenv("DB_FILE")
	serverName := os.Getenv("MATRIX_SERVER_NAME")
	username := os.Getenv("MATRIX_USERNAME")
	password := os.Getenv("MATRIX_PASSWORD")
	workflowsDefTOMLFile := os.Getenv("WORKFLOWS_DEF_TOML_FILE")

	logger.WithField("path", *envFile).Info("Loaded environment variables from .env")
	logger.Infof("Enabling debug? %t", debug)
	logger.Infof("Using database file: %s", dbFile)

	databaseSession, err := database.MakeDatabaseSession()
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"path": dbFile,
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

	botRegistry, err := makeBotRegistry(serverName, botRepository)
	if err != nil {
		logger.WithError(err).Fatal("Failed to make bot registry")
	}

	// set default port for running webhook listener server
	webhookListenerPort, err := strconv.Atoi(os.Getenv("WEBHOOK_LISTENER_PORT"))
	if err != nil {
		webhookListenerPort = 8080
	}
	webhookListenerServer := http.NewServer(webhookListenerPort)

	// if either one matrix related env var is specified, make sure all of them are specified
	if serverName != "" || username != "" || password != "" {
		if serverName == "" || username == "" || password == "" {
			logger.Fatalf("All matrix related variables need to be supplied if even one of them is supplied")
		}
	}

	// resolve .well-known to find our server URL to connect
	start := time.Now()
	serverURL := matrix.DiscoverServerURL(serverName)
	logger.WithFields(log.Fields{
		"serverName": serverName,
		"serverURL":  serverURL,
	}).WithDuration(time.Since(start)).Info("Discovered client API")

	p := engine.RunParams{
		BotRegistry:            botRegistry,
		WorkflowRepository:     workflowRepository,
		WorkflowStepRepository: workflowStepsRepository,
	}

	e := engine.NewEngine(p)
	e.StartUp()

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

	bots = append(bots, b.Bot{
		ID:          0,
		Description: "Primary bot",
		Username:    os.Getenv("MATRIX_USERNAME"),
		Password:    os.Getenv("MATRIX_PASSWORD"),
		Active:      true,
		Primary:     true,
	})

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
