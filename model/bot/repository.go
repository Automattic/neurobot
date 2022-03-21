package bot

// Repository facilitates persistence and retrieval of bots.
type Repository interface {
	// FindActive retrieves all active bots.
	FindActive() ([]Bot, error)

	// FindByIdentifier retrieves a bot by its unique identifier (aka slug).
	FindByIdentifier(identifier string) (Bot, error)
}
