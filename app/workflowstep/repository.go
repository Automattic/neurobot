package workflowstep

import (
	model "neurobot/model/workflowstep"

	"github.com/upper/db/v4"
)

type meta struct {
	ID     uint64 `db:"id,omitempty"`
	StepID uint64 `db:"step_id"`
	Key    string `db:"key"`
	Value  string `db:"value"`
}

type repository struct {
	collection     db.Collection
	collectionMeta db.Collection
}

func NewRepository(session db.Session) model.Repository {
	return &repository{
		collection:     session.Collection("workflow_step"),
		collectionMeta: session.Collection("workflow_step_meta"),
	}
}
