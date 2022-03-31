package workflow

// Repository facilitates persistence and retrieval of workflows.
type Repository interface {
	// Save persists a workflow
	Save(workflow *Workflow) error

	// Save identifier to workflow_meta table
	// would be removed once workflow meta table is removed
	SaveMeta(workflow *Workflow) error

	// FindActive retrieves all active workflows.
	FindActive() ([]Workflow, error)

	// FindByID retrieves a workflow by its ID.
	FindByID(ID uint64) (Workflow, error)

	// FindByIdentifier retrieves a workflow by its unique identifier.
	FindByIdentifier(identifier string) (Workflow, error)
}
