package workflow

type Workflow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Active      bool   `db:"active"`
	Identifier  string `db:"identifier"`
}
