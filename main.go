package main

import "matrix-workflow-builder/engine"

// @TODO all flag parsing and environment variables parsing is to come here in main.go

func main() {
	p := engine.RunParams{
		Debug:               true,
		PortWebhookListener: "8080",
	}

	e := engine.NewEngine(p)

	e.Startup()
	defer e.ShutDown()

	e.Run()
}
