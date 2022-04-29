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
	Run(wf.Workflow, payload.Payload) error
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

func (e *engine) Run(w wf.Workflow, payload payload.Payload) error {
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
			runners = append(runners, s.NewPostMatrixMessageRunner(step.Meta, e.botRegistry))
		case "stdOut":
			runners = append(runners, s.NewStdOutRunner(step.Meta, e.botRegistry))
		case "fetchDataExternal":
			runners = append(runners, s.NewFetchDataExternalRunner(step.Meta, e.botRegistry))
		}
	}

	for index, r := range runners {
		err = r.Run(&payload)
		if err != nil {
			// For now, we don't halt the workflow if a workflow step encounters an error
			logger.WithError(err).WithFields(log.Fields{
				"Identifier": w.Identifier,
			}).Info("workflow step execution error")
		}
		logger.WithFields(log.Fields{
			"index":   index,
			"payload": payload,
		}).Info("payload after step execution")
	}

	return nil
}
