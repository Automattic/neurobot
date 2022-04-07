package engine

import "github.com/apex/log"

type workflow struct {
	id          uint64
	name        string
	description string
	payload     map[string]string
	steps       []WorkflowStep
}

func (w *workflow) addWorkflowStep(s WorkflowStep) {
	w.steps = append(w.steps, s)
}

func (w *workflow) run(payload interface{}, e *engine) {
	logger := log.WithFields(log.Fields{
		"workflow": w.id,
		"payload":  payload,
	})
	logger.Info("Running workflow")

	// save payload inside of workflow, as we rinse-repeat it within the loop below
	w.payload = payload.(map[string]string)

	var err error
	// loop through all the steps inside of this workflow
	for _, s := range w.steps {
		w.payload, err = s.run(w.payload) // overwrite payload with each step execution and keep on passing this payload to each step
		if err != nil {
			// For now, we don't halt the workflow if a step encounters an error
			logger.WithError(err).Error("Workflow step execution error")
		}
	}
}
