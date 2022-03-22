package workflow

// Repository facilitates persistence and retrieval of workflows.
type Repository interface {
	// FindActive retrieves all active workflows.
	FindActive() ([]Workflow, error)
}
