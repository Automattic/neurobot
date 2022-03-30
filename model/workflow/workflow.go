package workflow

import "neurobot/model/workflowstep"

type Workflow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Active      bool   `db:"active"`
	Identifier  string
	Steps       []workflowstep.WorkflowStep
}
