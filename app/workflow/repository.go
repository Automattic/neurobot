package workflow

import (
	"github.com/upper/db/v4"
	model "neurobot/model/workflow"
)

const identifierKey = "toml_identifier"

type meta struct {
	ID         uint64 `db:"id,omitempty"`
	WorkflowID uint64 `db:"workflow_id"`
	Key        string `db:"key"`
	Value      string `db:"value"`
}

type repository struct {
	collection     db.Collection
	collectionMeta db.Collection
}

func NewRepository(session db.Session) model.Repository {
	return &repository{
		collection:     session.Collection("workflows"),
		collectionMeta: session.Collection("workflow_meta"),
	}
}

func (repository *repository) FindActive() (workflows []model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	if err := result.All(&workflows); err != nil {
		return nil, err
	}

	for index := range workflows {
		if err := repository.loadMeta(&workflows[index]); err != nil {
			return nil, err
		}
	}

	return
}

func (repository *repository) FindByID(ID uint64) (workflow model.Workflow, err error) {
	result := repository.collection.Find(db.Cond{"id": ID})
	err = result.One(&workflow)
	if err != nil {
		return
	}

	err = repository.loadMeta(&workflow)

	return
}

func (repository *repository) FindByIdentifier(identifier string) (workflow model.Workflow, err error) {
	var meta meta
	result := repository.collectionMeta.Find(db.Cond{"key": "toml_identifier", "value": identifier})
	err = result.One(&meta)

	workflow, err = repository.FindByID(meta.WorkflowID)

	return
}

// Load information from workflow_meta table into a workflow object.
func (repository *repository) loadMeta(workflow *model.Workflow) (err error) {
	var metas []meta
	result := repository.collectionMeta.Find(db.Cond{"workflow_id": workflow.ID})
	if err := result.All(&metas); err != nil {
		return err
	}

	for _, meta := range metas {
		if meta.Key == identifierKey {
			workflow.Identifier = meta.Value
		}
	}

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

	return result.Update(existing)
}

func (repository *repository) insert(workflow *model.Workflow) (err error) {
	result, err := repository.collection.Insert(workflow)
	if err == nil {
		workflow.ID = uint64(result.ID().(int64))
	}

	return
}
