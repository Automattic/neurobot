package engine

import (
	wf "neurobot/model/workflow"
)

// get all active workflows out of the database
func getConfiguredWorkflows(repository wf.Repository) (w []wf.Workflow, err error) {
	savedWorkflows, err := repository.FindActive()
	if err != nil {
		return
	}

	for _, row := range savedWorkflows {
		w = append(w, wf.Workflow{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
		})
	}

	return
}
