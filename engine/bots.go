package engine

import (
	"github.com/upper/db/v4"
)

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

func getBot(dbs db.Session, identifier string) (b Bot, err error) {
	res := dbs.Collection("bots").Find(db.Cond{"identifier": identifier})
	err = res.One(&b)

	return
}
