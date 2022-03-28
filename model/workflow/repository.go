package workflow

// Repository facilitates persistence and retrieval of workflows.
type Repository interface {
	// Save persists a workflow
	Save(workflow *Workflow) error

	// FindActive retrieves all active workflows.
	FindActive() ([]Workflow, error)

	// FindByID retrieves a workflow by its ID.
	FindByID(ID uint64) (Workflow, error)

	// FindByIdentifier retrieves a workflow by its unique identifier.
	FindByIdentifier(identifier string) (Workflow, error)

	// GetTOMLMapping returns mapping of toml identifiers with their respective database IDs
	GetTOMLMapping() (map[string]uint64, error)
}
