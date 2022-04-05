package main

import (
	"flag"
	"fmt"
	"log"
	application "neurobot/app"
	botApp "neurobot/app/bot"
	"neurobot/app/workflow"
	"neurobot/app/workflowstep"
	"neurobot/infrastructure/database"
	"neurobot/infrastructure/http"
	"neurobot/infrastructure/matrix"
	"neurobot/infrastructure/toml"
	b "neurobot/model/bot"
	"os"
	"strconv"
	"strings"

	"neurobot/engine"

	"github.com/joho/godotenv"
	"maunium.net/go/mautrix"
)

var envFile = flag.String("env", "./.env", ".env file")

func main() {
	flag.Parse()

	err := godotenv.Load(*envFile)
	if err != nil {
		log.Fatalf("Error loading .env file at %s. Err: %s\n", *envFile, err)
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false // default
	}

	dbFile := os.Getenv("DB_FILE")
	serverName := os.Getenv("MATRIX_SERVER_NAME")
	username := os.Getenv("MATRIX_USERNAME")
	password := os.Getenv("MATRIX_PASSWORD")
	workflowsDefTOMLFile := os.Getenv("WORKFLOWS_DEF_TOML_FILE")

	log.Println("Debug:", debug)
	log.Printf("Loaded environment variables from %s\n", *envFile)
	log.Printf("Using database file %s\n", dbFile)

	databaseSession, err := database.MakeDatabaseSession()
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer databaseSession.Close()
	err = database.Migrate(databaseSession)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	botRepository := botApp.NewRepository(databaseSession)
	workflowRepository := workflow.NewRepository(databaseSession)
	workflowStepsRepository := workflowstep.NewRepository(databaseSession)

	// import TOML
	err = toml.Import(workflowsDefTOMLFile, workflowRepository, workflowStepsRepository)
	if err != nil {
		log.Fatalf("error while importing TOML workflows: %s", err)
	}

	botRegistry, err := makeBotRegistry(serverName, botRepository)
	if err != nil {
		log.Fatalf("%s", err)
	}

	// set default port for running webhook listener server
	webhookListenerPort, err := strconv.Atoi(os.Getenv("WEBHOOK_LISTENER_PORT"))
	if err != nil {
		webhookListenerPort = 8080
	}
	webhookListenerServer := http.NewServer(webhookListenerPort)

	// if either one matrix related env var is specified, make sure all of them are specified
	isMatrix := false
	if serverName != "" || username != "" || password != "" {
		if serverName == "" || username == "" || password == "" {
			log.Fatalf("All matrix related variables need to be supplied if even one of them is supplied")
		} else {
			isMatrix = true
		}
	}

	// resolve .well-known to find our server URL to connect
	var serverURL string
	log.Printf("Discovering Client API for %s\n", serverName)
	wellKnown, err := mautrix.DiscoverClientAPI(serverName) // both can be nil for hosts that have https but are not a matrix server
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
			serverURL = "http://" + serverName
		} else {
			serverURL = "https://" + serverName
		}
	} else {
		if wellKnown != nil {
			serverURL = wellKnown.Homeserver.BaseURL
		} else {
			serverURL = "https://" + serverName
		}
	}
	log.Printf("Server URL for %s: %s", serverName, serverURL)

	p := engine.RunParams{
		BotRepository:          botRepository,
		BotRegistry:            botRegistry,
		WorkflowRepository:     workflowRepository,
		WorkflowStepRepository: workflowStepsRepository,

		Debug:            debug,
		IsMatrix:         isMatrix,
		MatrixServerName: serverName,
		MatrixServerURL:  serverURL,
		MatrixUsername:   username,
		MatrixPassword:   password,
	}

	e := engine.NewEngine(p)

	if isMatrix {
		mc, err := mautrix.NewClient(p.MatrixServerURL, "", "")
		if err != nil {
			log.Fatal(err)
		}
		e.StartUp(mc, mc.Syncer.(*mautrix.DefaultSyncer))
	} else {
		fmt.Println("engine:", "Lite mode")
		e.StartUpLite()
	}

	app := application.NewApp(e, botRegistry, workflowRepository, webhookListenerServer)
	if err := app.Run(); err != nil {
		log.Fatalf("%s", err)
	}

	log.Printf("Starting webhook listener at port %d\n", webhookListenerPort)
	webhookListenerServer.Run() // blocking
}

func makeBotRegistry(homeserverURL string, botRepository b.Repository) (registry botApp.Registry, err error) {
	bots, err := botRepository.FindActive()
	if err != nil {
		return
	}

	bots = append(bots, b.Bot{
		ID:          0,
		Identifier:  os.Getenv("MATRIX_USERNAME"),
		Name:        "Neurobot",
		Description: "Primary bot",
		Username:    os.Getenv("MATRIX_USERNAME"),
		Password:    os.Getenv("MATRIX_PASSWORD"),
		Active:      true,
		Primary:     true,
	})

	registry = botApp.NewRegistry(homeserverURL)

	for _, bot := range bots {
		client, err := matrix.NewMautrixClient(homeserverURL, true)
		if err == nil {
			err = registry.Append(bot, client)
		}
	}

	return
}
