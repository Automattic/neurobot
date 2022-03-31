package app

import (
	"fmt"
	netHttp "net/http"
	"neurobot/app/bot"
	r "neurobot/app/runner"
	"neurobot/engine"
	"neurobot/infrastructure/event"
	"neurobot/infrastructure/http"
	b "neurobot/model/bot"
	w "neurobot/model/workflow"
	"strings"
)

type app struct {
	engine             engine.Engine
	eventBus           event.Bus
	botRepository      b.Repository
	botRegistry        bot.Registry
	workflowRepository w.Repository
	webhookListener    *http.Server
}

func NewApp(
	engine engine.Engine,
	eventBus event.Bus,
	botRegistry bot.Registry,
	workflowRepository w.Repository,
	webhookListener *http.Server,
) *app {
	return &app{
		engine:             engine,
		eventBus:           eventBus,
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
				netHttp.NotFound(response, request)
				return
			}

			go func() {
				err := app.runWorkflow(workflow, payload)
				if err != nil {
					fmt.Printf("Failed to run workflow: %s", err)
				}
			}()
		})

	return
}

func (app app) runWorkflow(workflow w.Workflow, payload map[string]string) error {
	var runner r.Runner

	switch workflow.Identifier {
	default:
		runner = app.engine
	}

	return runner.Run(workflow, payload)
}
