package workflow

import (
	"github.com/upper/db/v4"
	model "neurobot/model/workflow"
)

type Repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) *Repository {
	return &Repository{
		collection: session.Collection("workflows"),
	}
}

func (repository *Repository) FindActive() (bots []model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	err = result.All(&bots)
	return
}
