package engine

import (
	"fmt"
)

type workflow struct {
	id          uint64
	name        string
	description string
	payload     interface{}
	steps       []WorkflowStep
}

func (w *workflow) addWorkflowStep(s WorkflowStep) {
	w.steps = append(w.steps, s)
}

func (w *workflow) run(payload string) {
	fmt.Printf("\nRunning workflow #%d\n", w.id)
	// loop through all the steps inside of this workflow
	for _, s := range w.steps {
		w.payload = s.run(payload) // overwrite payload with each step execution and keep on passing this payload to each step
	}
}
