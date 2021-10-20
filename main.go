package main

import (
	"flag"
	"fmt"
	"log"
	"matrix-workflow-builder/engine"
	"os"
	"strconv"

	"maunium.net/go/mautrix"
)

var homeserver = flag.String("homeserver", "", "Matrix Homeserver URL")
var username = flag.String("username", "", "Matrix username localpart")
var password = flag.String("password", "", "Matrix password")
var debug = flag.String("debug", "false", "Debug mode")
var dbFile = flag.String("dbfile", "./wfb.db", "Database file")
var webhookListenerPort = flag.String("webhooklistenerport", "8080", "Webhook Listener Port")

func main() {
	flag.Parse()
	if *username == "" || *password == "" || *homeserver == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	debug, err := strconv.ParseBool(*debug)
	if err != nil {
		debug = false
	}

	fmt.Println("debug:", debug)

	p := engine.RunParams{
		Debug:               debug,
		Database:            *dbFile,
		PortWebhookListener: *webhookListenerPort,
		MatrixHomeServer:    *homeserver,
		MatrixUsername:      *username,
		MatrixPassword:      *password,
	}

	e := engine.NewEngine(p)

	mc, err := mautrix.NewClient(*homeserver, "", "")
	if err != nil {
		log.Fatal(err)
	}

	e.StartUp(mc, mc.Syncer.(*mautrix.DefaultSyncer))
	defer e.ShutDown()

	e.Run()
}
