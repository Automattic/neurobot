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

func (repository *repository) Save(step *model.WorkflowStep) error {
	if step.ID > 0 {
		return repository.update(step)
	}

	return repository.insert(step)
}

func (repository *repository) insert(step *model.WorkflowStep) (err error) {
	result, err := repository.collection.Insert(step)
	if err != nil {
		return
	}

	step.ID = uint64(result.ID().(int64))

	repository.saveMeta(step)

	return
}

func (repository *repository) update(step *model.WorkflowStep) (err error) {
	var existing model.WorkflowStep

	result := repository.collection.Find(step.ID)
	if err = result.One(&existing); err != nil {
		return
	}

	existing.Name = step.Name
	existing.Description = step.Description
	existing.Active = step.Active
	existing.Variety = step.Variety
	existing.SortOrder = step.SortOrder

	if err = result.Update(existing); err != nil {
		return
	}

	repository.saveMeta(step)

	return
}

// Load information from workflow_meta table into a workflow object.
func (repository *repository) loadMeta(step *model.WorkflowStep) (err error) {
	var metas []meta
	result := repository.collectionMeta.Find(db.Cond{"step_id": step.ID})
	if err := result.All(&metas); err != nil {
		return err
	}

	for _, meta := range metas {
		step.Meta[meta.Key] = meta.Value
	}

	return
}

func (repository *repository) saveMeta(step *model.WorkflowStep) (err error) {
	// delete existing entries
	res := repository.collectionMeta.Find(db.Cond{"step_id": step.ID})
	if err = res.Delete(); err != nil {
		return
	}

	// insert new entries
	for k, v := range step.Meta {
		_, err = repository.collectionMeta.Insert(meta{
			StepID: step.ID,
			Key:    k,
			Value:  v,
		})
		if err != nil {
			return
		}
	}

	return
}
