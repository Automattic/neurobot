package engine

import (
	wf "neurobot/model/workflow"
	wfs "neurobot/model/workflowstep"
)

// get all active workflows out of the database
func getConfiguredWorkflows(repository wf.Repository) (w []workflow, err error) {
	savedWorkflows, err := repository.FindActive()
	if err != nil {
		return
	}

	for _, row := range savedWorkflows {
		w = append(w, workflow{
			id:          row.ID,
			name:        row.Name,
			description: row.Description,
		})
	}

	return
}

// get all active workflow steps out of the database
func getConfiguredWFSteps(repository wfs.Repository) (s []WorkflowStep, err error) {
	savedSteps, err := repository.FindActive()
	if err != nil {
		return
	}

	// range over all active steps, collecting meta for each step and appending that to collect basket
	for _, step := range savedSteps {
		switch step.Variety {
		case "postMatrixMessage":
			s = append(s, &postMessageMatrixWorkflowStep{
				workflowStep: workflowStep{
					id:          step.ID,
					name:        step.Name,
					description: step.Description,
					variety:     step.Variety,
					workflowID:  step.WorkflowID,
				},
				postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
					messagePrefix: step.Meta["messagePrefix"],
					room:          step.Meta["matrixRoom"],
					asBot:         step.Meta["asBot"],
				},
			})
		case "stdout":
			s = append(s, &stdoutWorkflowStep{
				workflowStep: workflowStep{
					id:          step.ID,
					name:        step.Name,
					description: step.Description,
					variety:     step.Variety,
					workflowID:  step.WorkflowID,
				},
			})
		}
	}

	return
}
