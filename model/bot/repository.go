package bot

// Repository facilitates persistence and retrieval of bots.
type Repository interface {
	// Save persists a bot.
	Save(bot *Bot) error

	// FindActive retrieves all active bots.
	FindActive() ([]Bot, error)

	// FindByUsername retrieves a bot by its unique username.
	FindByUsername(username string) (Bot, error)
}
