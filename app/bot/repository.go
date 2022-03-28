package bot

import (
	"github.com/upper/db/v4"
	model "neurobot/model/bot"
)

type repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) model.Repository {
	return &repository{
		collection: session.Collection("bots"),
	}
}

func (repository *repository) FindActive() (bots []model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	err = result.All(&bots)
	return
}

func (repository *repository) FindByIdentifier(identifier string) (bot model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"identifier": identifier})
	err = result.One(&bot)
	return
}
