package engine

import (
	"fmt"
	"neurobot/app/bot"
	wf "neurobot/model/workflow"
	wfs "neurobot/model/workflowstep"

	"github.com/apex/log"

	"github.com/upper/db/v4"
)

type Engine interface {
	StartUp()
	Run(wf.Workflow, map[string]string) error
}

type WorkflowStepRunner interface {
	run(map[string]string) (map[string]string, error) // accepts payload and returns after modification (if desired)
}

type engine struct {
	debug bool

	db        db.Session
	workflows map[uint64]*wf.Workflow

	botRegistry            bot.Registry
	workflowRepository     wf.Repository
	workflowStepRepository wfs.Repository
}

type RunParams struct {
	Debug bool

	BotRegistry            bot.Registry
	WorkflowRepository     wf.Repository
	WorkflowStepRepository wfs.Repository
}

func (e *engine) Run(w wf.Workflow, payload map[string]string) error {
	logger := log.Log

	// loop through all the steps inside of the workflow
	steps, err := e.workflowStepRepository.FindByWorkflowID(w.ID)
	if err != nil {
		return fmt.Errorf("error fetching workflow steps while running workflow %d : %w", w.ID, err)
	}

	var runners []WorkflowStepRunner

	for _, s := range steps {
		switch s.Variety {
		case "postMatrixMessage":
			runners = append(runners, NewPostMatrixMessageRunner(s.Meta, e.botRegistry))
		case "stdOut":
			runners = append(runners, NewStdOutRunner(s.Meta, e.botRegistry))
		}
	}

	for _, r := range runners {
		payload, err = r.run(payload)
		if err != nil {
			// For now, we don't halt the workflow if a workflow step encounters an error
			logger.WithError(err).WithFields(log.Fields{
				"Identifier": w.Identifier,
			}).Info("workflow step execution error")
		}
	}

	return nil
}

func (e *engine) StartUp() {
	logger := log.Log
	logger.Info("Starting up engine")

	workflows, err := e.workflowRepository.FindActive()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load workflows from database")
	}
	for _, w := range workflows {
		// copy over the value in a separate variable because we need to store a pointer
		// w gets assigned a different value with every iteration, which modifies all values if address of w is taken directly
		instance := w
		e.workflows[w.ID] = &instance
	}

	logger.Info("Finished starting up engine.")
}

func NewEngine(p RunParams) *engine {
	e := engine{}

	// setting run parameters
	e.debug = p.Debug
	e.botRegistry = p.BotRegistry
	e.workflowRepository = p.WorkflowRepository
	e.workflowStepRepository = p.WorkflowStepRepository

	// initialize maps
	e.workflows = make(map[uint64]*wf.Workflow)

	return &e
}
