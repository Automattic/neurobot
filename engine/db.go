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
