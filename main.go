package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

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
		return
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false // default
	}

	dbFile := os.Getenv("DB_FILE")
	homeserver := os.Getenv("MATRIX_HOMESERVER")
	username := os.Getenv("MATRIX_USERNAME")
	password := os.Getenv("MATRIX_PASSWORD")
	webhookListenerPort := os.Getenv("WEBHOOK_LISTENER_PORT")
	workflowsDefTOMLFile := os.Getenv("WORKFLOWS_DEF_TOML_FILE")

	if debug {
		log.Println("Debug:", debug)
		fmt.Printf("Loaded environment variables from %s\n", *envFile)
		fmt.Printf("Using database file %s\n", dbFile)
	}

	// if either one matrix related env var is specified, make sure all of them are specified
	isMatrix := false
	if homeserver != "" || username != "" || password != "" {
		if homeserver == "" || username == "" || password == "" {
			log.Fatalf("All matrix related variables need to be supplied if even one of them is supplied")
		} else {
			isMatrix = true
		}
	}

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
		MatrixHomeServer:     homeserver,
		MatrixUsername:       username,
		MatrixPassword:       password,
	}

	e := engine.NewEngine(p)

	if isMatrix {
		mc, err := mautrix.NewClient(homeserver, "", "")
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
