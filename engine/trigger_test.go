package engine

import (
	"testing"
)

// This function works on the idea of defining the impact of running a workflow
// step as making a small change to the original payload.
// After trigger invokes the workflow, all workflow steps should have changed the
// original payload in a specific way, so just examine the final payload.
func TestTriggerFinish(t *testing.T) {
	tables := []struct {
		initialPayload   string
		impact           string
		processedPayload string
	}{
		{
			initialPayload:   "thor, son of ",
			impact:           "odin",
			processedPayload: "thor, son of odin",
		},
		{
			initialPayload:   "thor, son of",
			impact:           " odin",
			processedPayload: "thor, son of odin",
		},
	}

	for _, table := range tables {
		e := Engine{}
		e.workflows = make(map[uint64]*workflow)

		tg := trigger{
			engine:    &e,
			workflows: []uint64{1},
		}

		w := &workflow{id: 1, steps: []WorkflowStep{}}
		w.addWorkflowStep(NewMockWorkflowStep(
			table.impact,
		))
		e.workflows[1] = w

		tg.finish(table.initialPayload)

		if w.payload != table.processedPayload {
			t.Errorf("trigger finish didn't generate the right processed payload. expected: '%s' got: '%s'", table.processedPayload, w.payload)
		}
	}
}
