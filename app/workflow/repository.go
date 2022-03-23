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

func NewRepository(session db.Session) *repository {
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
