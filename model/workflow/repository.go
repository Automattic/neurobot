package workflow

// Repository facilitates persistence and retrieval of workflows.
type Repository interface {
	// Save persists a workflow
	Save(workflow *Workflow) error

	// FindActive retrieves all active workflows.
	FindActive() ([]Workflow, error)

	// FindByID retrieves a workflow by its ID.
	FindByID(ID uint64) (Workflow, error)
}
