package app

import (
	"fmt"
	netHttp "net/http"
	"neurobot/app/bot"
	"neurobot/app/engine"
	r "neurobot/app/runner"
	"neurobot/app/runner/afk_notifier"
	"neurobot/infrastructure/http"
	w "neurobot/model/workflow"
	"strings"

	"github.com/apex/log"
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
				log.WithError(err).WithFields(log.Fields{
					"payload": payload,
				}).Error("failed to run workflow")
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
		log.WithFields(log.Fields{
			"identifier": workflow.Identifier,
			"payload":    payload,
		}).Info("starting workflow")
		err := runner.Run(workflow, payload)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"payload": payload,
			}).Error("failed to run workflow")
		}
	}()

	return nil
}
