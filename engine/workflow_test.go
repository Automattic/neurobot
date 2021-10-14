package engine

import "testing"

func TestAddWorkflowStep(t *testing.T) {
	w := &workflow{steps: []WorkflowStep{}}
	w.addWorkflowStep(NewMockWorkflowStep("g"))
	if len(w.steps) != 1 {
		t.Errorf("workflow step %s wasn't added", "g")
	}
	w.addWorkflowStep(NewMockWorkflowStep("o"))
	if len(w.steps) != 2 {
		t.Errorf("workflow step %s wasn't added", "o")
	}
}

// This function works on the idea of defining the impact of running a workflow
// step as making a small change to the original payload.
// After running a workflow, all workflow steps should have changed the
// original payload in a specific way, so just examine the final payload.
func TestRun(t *testing.T) {
	tables := []struct {
		triggerPayload  string
		expectedPayload string
		impacts         []string // change in payload as a proof of that workflowstep's execution
	}{
		{
			triggerPayload:  "love",
			impacts:         []string{" ", "m", "a", "t", "r", "i", "x"},
			expectedPayload: "love matrix",
		},
		{
			triggerPayload:  "",
			impacts:         []string{"g", "o"},
			expectedPayload: "go",
		},
		{
			triggerPayload:  "love",
			impacts:         []string{},
			expectedPayload: "love",
		},
	}
	for _, table := range tables {
		w := &workflow{payload: table.triggerPayload, steps: []WorkflowStep{}}
		for _, i := range table.impacts {
			w.addWorkflowStep(NewMockWorkflowStep(i))
		}

		w.run(w.payload, &Engine{})

		if w.payload != table.expectedPayload {
			t.Errorf("workflow ran workflow steps but final payload was '%s', expected: '%s'", w.payload, table.expectedPayload)
		}
	}
}
