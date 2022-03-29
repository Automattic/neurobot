package workflowstep

type WorkflowStep struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Variety     string `db:"variety"`
	WorkflowID  uint64 `db:"workflow_id"`
	SortOrder   uint64 `db:"sort_order"`
	Active      int    `db:"active"`
}
