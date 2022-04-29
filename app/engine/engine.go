package engine

import (
	"fmt"
	"neurobot/app/bot"
	s "neurobot/app/engine/steps"
	"neurobot/model/payload"
	wf "neurobot/model/workflow"
	wfs "neurobot/model/workflowstep"

	"github.com/apex/log"
)

type Engine interface {
	Run(string, wf.Workflow, payload.Payload) error
}

type WorkflowStepRunner interface {
	Run(*payload.Payload) error // accepts payload pointer (for easy modification)
}

type engine struct {
	botRegistry            bot.Registry
	workflowStepRepository wfs.Repository
}

func NewEngine(botRegistry bot.Registry, workflowStepRepository wfs.Repository) *engine {
	return &engine{
		botRegistry:            botRegistry,
		workflowStepRepository: workflowStepRepository,
	}
}

func (e *engine) Run(eid string, w wf.Workflow, payload payload.Payload) error {
	logger := log.Log

	// loop through all the steps inside of the workflow
	steps, err := e.workflowStepRepository.FindByWorkflowID(w.ID)
	if err != nil {
		return fmt.Errorf("error fetching workflow steps while running workflow %d : %w", w.ID, err)
	}

	var runners []WorkflowStepRunner

	for _, step := range steps {
		switch step.Variety {
		case "postMatrixMessage":
			runners = append(runners, s.NewPostMatrixMessageRunner(eid, step.Meta, e.botRegistry))
		case "stdOut":
			runners = append(runners, s.NewStdOutRunner(eid, step.Meta))
		case "fetchDataExternal":
			runners = append(runners, s.NewFetchDataExternalRunner(eid, step.Meta))
		case "formatMessage":
			runners = append(runners, s.NewFormatMessageRunner(eid, step.Meta))
		}
	}

	for index, r := range runners {
		ctx := log.Fields{
			"executionID": eid,
			"identifier":  w.Identifier,
			"index":       index,
			"payload":     payload,
		}
		err = r.Run(&payload)
		if err != nil {
			// For now, we don't halt the workflow if a workflow step encounters an error
			logger.WithError(err).WithFields(ctx).Info("workflow step execution error")
		}
		logger.WithFields(ctx).Debug("payload after step execution")
	}

	return nil
}
