package main

import (
	"flag"
	"fmt"
	"log"
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
	webhookListenerPort := os.Getenv("WEBHOOK_LISTENER_PORT")
	workflowsDefTOMLFile := os.Getenv("WORKFLOWS_DEF_TOML_FILE")

	log.Println("Debug:", debug)
	log.Printf("Loaded environment variables from %s\n", *envFile)
	log.Printf("Using database file %s\n", dbFile)

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
	wellKnown, err := mautrix.DiscoverClientAPI(serverName)
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

	// set default port for running webhook listener server
	if webhookListenerPort == "" {
		webhookListenerPort = "8080"
	}

	p := engine.RunParams{
		Debug:                debug,
		Database:             dbFile,
		PortWebhookListener:  webhookListenerPort,
		WorkflowsDefTOMLFile: workflowsDefTOMLFile,
		IsMatrix:             isMatrix,
		MatrixServerName:     serverName,
		MatrixServerURL:      serverURL,
		MatrixUsername:       username,
		MatrixPassword:       password,
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
	defer e.ShutDown()

	e.Run()
}
