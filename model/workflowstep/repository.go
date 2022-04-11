package workflowstep

// Repository facilitates persistence and retrieval of workflow steps.
type Repository interface {
	// Save persists a workflow step
	Save(workflowStep *WorkflowStep) error

	// FindActive retrieves all active workflow steps
	FindActive() ([]WorkflowStep, error)

	// FindByID retrieves a workflow step by its ID.
	FindByID(ID uint64) (WorkflowStep, error)

	// Find all workflow steps by workflow ID
	FindByWorkflowID(ID uint64) ([]WorkflowStep, error)

	// Removes all workflow steps under a workflow
	RemoveByWorkflowID(ID uint64) error
}
