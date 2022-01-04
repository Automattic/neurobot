package main

import (
	"flag"
	"fmt"
	"log"
	"matrix-workflow-builder/engine"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"maunium.net/go/mautrix"
)

var debug = flag.String("debug", "false", "Debug mode")
var envFile = flag.String("env", "./.env", ".env file")
var dbFile = flag.String("dbfile", "./wfb.db", "Database file")

func main() {
	flag.Parse()

	debug, err := strconv.ParseBool(*debug)
	if err != nil {
		debug = false
	}

	fmt.Println("debug:", debug)

	err = godotenv.Load(*envFile)
	if err != nil {
		log.Fatalf("Error loading .env file at %s. Err: %s\n", *envFile, err)
	}

	homeserver := os.Getenv("MATRIX_HOMESERVER")
	username := os.Getenv("MATRIX_USERNAME")
	password := os.Getenv("MATRIX_PASSWORD")
	webhookListenerPort := os.Getenv("WEBHOOK_LISTENER_PORT")
	staticServerPort := os.Getenv("STATIC_SERVER_PORT")

	// if either one matrix related env var is specified, make sure all of them are specified
	isMatrix := false
	if homeserver != "" || username != "" || password != "" {
		if homeserver == "" || username == "" || password == "" {
			log.Fatalf("All matrix related variables need to be supplied if even one of them is supplied")
		} else {
			isMatrix = true
		}
	}

	// set default ports for running servers, if not provided
	if webhookListenerPort == "" {
		webhookListenerPort = "8080"
	}
	if staticServerPort == "" {
		staticServerPort = "8181"
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run static server for UI
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(http.Dir("./ui")))

		log.Println("Starting static file server on :" + staticServerPort)
		staticServer := http.Server{
			Addr:    fmt.Sprintf(":%v", staticServerPort),
			Handler: mux,
		}
		log.Fatal(staticServer.ListenAndServe())
		wg.Done()
	}()

	// Run Workflow builder engine
	go func() {
		p := engine.RunParams{
			Debug:               debug,
			Database:            *dbFile,
			PortWebhookListener: webhookListenerPort,
			IsMatrix:            isMatrix,
			MatrixHomeServer:    homeserver,
			MatrixUsername:      username,
			MatrixPassword:      password,
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
		wg.Done()
	}()

	wg.Wait()
}
