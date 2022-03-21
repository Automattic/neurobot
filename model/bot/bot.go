package bot

// Bot is a non-human chat user.
type Bot struct {
	ID          uint64 `db:"id,omitempty"`
	Identifier  string `db:"identifier"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	CreatedBy   string `db:"created_by"`
	Active      bool   `db:"active"`
}
