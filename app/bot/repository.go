package bot

import (
	"github.com/upper/db/v4"
	model "neurobot/model/bot"
)

type Repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) *Repository {
	return &Repository{
		collection: session.Collection("bots"),
	}
}

func (repository *Repository) FindActive() (bots []model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	err = result.All(&bots)
	return
}

func (repository *Repository) FindByIdentifier(identifier string) (bot model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"identifier": identifier})
	err = result.One(&bot)
	return
}
