package workflow

import (
	model "neurobot/model/workflow"

	"github.com/upper/db/v4"
)

type repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) model.Repository {
	return &repository{
		collection: session.Collection("workflows"),
	}
}

func (repository *repository) FindActive() (workflows []model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	if err := result.All(&workflows); err != nil {
		return nil, err
	}

	return
}

func (repository *repository) FindByID(ID uint64) (workflow model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"id": ID})
	err = result.One(&workflow)

	return
}

func (repository *repository) FindByIdentifier(identifier string) (workflow model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"identifier": identifier})
	err = result.One(&workflow)

	return
}

func (repository *repository) Save(workflow *model.Workflow) error {
	if workflow.ID > 0 {
		return repository.update(workflow)
	}

	return repository.insert(workflow)
}

func (repository *repository) update(workflow *model.Workflow) (err error) {
	var existing model.Workflow

	result := repository.collection.Find(workflow.ID)
	if err = result.One(&existing); err != nil {
		return
	}

	existing.Name = workflow.Name
	existing.Description = workflow.Description
	existing.Active = workflow.Active
	existing.Identifier = workflow.Identifier

	err = result.Update(existing)

	return
}

func (repository *repository) insert(workflow *model.Workflow) (err error) {
	result, err := repository.collection.Insert(workflow)
	if err == nil {
		workflow.ID = uint64(result.ID().(int64))
	}

	return
}
