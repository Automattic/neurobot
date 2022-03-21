package bot

import (
	"github.com/upper/db/v4"
)

type Repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) *Repository {
	return &Repository{
		collection: session.Collection("bots"),
	}
}
