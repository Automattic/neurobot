package bot

// Bot is a non-human chat user.
type Bot struct {
	ID          uint64 `db:"id,omitempty"`
	Description string `db:"description"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	Active      bool   `db:"active"`
	Primary     bool
}
