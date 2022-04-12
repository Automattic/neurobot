package app

import (
	"fmt"
	"log"
	netHttp "net/http"
	"neurobot/app/bot"
	"neurobot/app/engine"
	r "neurobot/app/runner"
	"neurobot/app/runner/afk_notifier"
	"neurobot/infrastructure/http"
	w "neurobot/model/workflow"
	"strings"
)

type app struct {
	engine             engine.Engine
	botRegistry        bot.Registry
	workflowRepository w.Repository
	webhookListener    *http.Server
}

func NewApp(
	engine engine.Engine,
	botRegistry bot.Registry,
	workflowRepository w.Repository,
	webhookListener *http.Server,
) *app {
	return &app{
		engine:             engine,
		botRegistry:        botRegistry,
		workflowRepository: workflowRepository,
		webhookListener:    webhookListener,
	}
}

func (app app) Run() (err error) {
	err = app.webhookListener.RegisterRoute(
		"/",
		func(response netHttp.ResponseWriter, request *netHttp.Request, payload map[string]string) {
			workflowIdentifier := strings.TrimPrefix(request.URL.Path, "/")
			workflow, err := app.workflowRepository.FindByIdentifier(workflowIdentifier)
			if err != nil {
				errorMessage := fmt.Sprintf("no workflow found for `%s`", workflowIdentifier)
				netHttp.Error(response, errorMessage, netHttp.StatusNotFound)
				return
			}

			err = app.runWorkflow(workflow, payload)
			if err != nil {
				netHttp.Error(response, "something went wrong", netHttp.StatusInternalServerError)
				log.Printf("Error when attempting to run workflow: %s, payload: %+v", err, payload)
				return
			}
		})

	return
}

func (app app) runWorkflow(workflow w.Workflow, payload map[string]string) error {
	var runner r.Runner

	switch workflow.Identifier {
	default:
		runner = app.engine
	case "afk_notifier":
		matrixClient, err := app.botRegistry.GetClient("afk")
		if err != nil {
			return err
		}
		runner = afk_notifier.NewRunner(matrixClient)
	}

	go func() {
		log.Printf("Starting workflow with identifier %s, payload: %+v", workflow.Identifier, payload)
		err := runner.Run(workflow, payload)
		if err != nil {
			log.Printf("Failed to run workflow: %s", err)
		}
	}()

	return nil
}
