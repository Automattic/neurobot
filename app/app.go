package app

import (
	"fmt"
	netHttp "net/http"
	"neurobot/app/bot"
	"neurobot/app/commands"
	"neurobot/app/engine"
	r "neurobot/app/runner"
	"neurobot/app/runner/afk_notifier"
	"neurobot/infrastructure/http"
	"neurobot/model/command"
	"neurobot/model/message"
	"neurobot/model/payload"
	"neurobot/model/room"
	w "neurobot/model/workflow"
	"strings"

	"github.com/apex/log"
	"github.com/google/uuid"
)

type app struct {
	engine             engine.Engine
	botRegistry        bot.Registry
	workflowRepository w.Repository
	webhookListener    *http.Server
	commandChannel     <-chan *command.Command
}

// NewApp returns the instance to run the entire program
func NewApp(
	engine engine.Engine,
	botRegistry bot.Registry,
	workflowRepository w.Repository,
	webhookListener *http.Server,
	commandChannel <-chan *command.Command,
) *app {
	return &app{
		engine:             engine,
		botRegistry:        botRegistry,
		workflowRepository: workflowRepository,
		webhookListener:    webhookListener,
		commandChannel:     commandChannel,
	}
}

func (app app) Run() (err error) {
	err = app.webhookListener.RegisterRoute(
		"/",
		func(response netHttp.ResponseWriter, request *netHttp.Request, payload payload.Payload) {
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
		},
	)

	go func() {
		log.WithFields(log.Fields{"port": app.webhookListener.Port()}).Info("Starting webhook listener")
		app.webhookListener.Run() // blocking
	}()

	// loop over commands as they are invoked and run each of them (blocking since we are looping over a channel)
	for c := range app.commandChannel {
		log.WithFields(log.Fields{
			"command": c.Name,
			"args":    c.Args,
			"room":    c.Meta["room"],
		}).Info("command received")

		go app.runCommand(c)
	}

	return
}

func (app app) runWorkflow(workflow w.Workflow, payload payload.Payload) error {
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
		// generate a UUID for this execution of the workflow
		eid := uuid.New()
		ctx := log.Fields{
			"executionID": eid.String(),
			"identifier":  workflow.Identifier,
			"payload":     payload,
		}
		log.WithFields(ctx).Info("starting workflow")
		err := runner.Run(eid.String(), workflow, payload)
		if err != nil {
			log.WithError(err).WithFields(ctx).Error("failed to run workflow")
		}
		log.WithFields(ctx).Info("finished workflow")
	}()

	return nil
}

func (app app) runCommand(comm *command.Command) {
	identifier := "COMMAND_" + strings.ToUpper(comm.Name)

	// create an instance of the command interface
	var command commands.Command
	switch strings.ToUpper(comm.Name) {
	default:
		command = commands.NewUnrecognized(comm)
	case "ECHO":
		command = commands.NewEcho(comm)
	case "POLYGLOTS":
		command = commands.NewPolyglots(comm)
	}

	payload := command.WorkflowPayload() // need payload to access room, even if command is not valid

	if !command.Valid() {
		payload.Message = command.UsageHints()

		mc, err := app.botRegistry.GetPrimaryClient()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"identifier": identifier,
			}).Error("responding to invalid command failed")
			return
		}

		roomID, _ := room.NewID(payload.Room) // no need to check for error, is picked up from an event and not a user input

		if err := mc.SendMessage(
			roomID,
			message.NewMarkdownMessage(command.UsageHints()),
		); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"identifier": identifier,
			}).Error("responding to invalid command failed")
			return
		}
	}

	// find workflow associated with this command
	workflow, err := app.workflowRepository.FindByIdentifier(identifier)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"identifier": identifier}).Error("no workflow found")
		return
	}

	// run workflow
	err = app.runWorkflow(workflow, payload)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"identifier": identifier,
			"payload":    payload,
		}).Error("failed to run workflow")
	}
}
