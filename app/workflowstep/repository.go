package workflowstep

import (
	"fmt"
	model "neurobot/model/workflowstep"
	"strings"

	"github.com/upper/db/v4"
)

const workflowStepTableName = "workflow_steps"
const workflowStepMetaTableName = "workflow_step_meta"

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
		collection:     session.Collection(workflowStepTableName),
		collectionMeta: session.Collection(workflowStepMetaTableName),
	}
}

// FindActive returns all active workflow steps from the database
func (repository *repository) FindActive() (steps []model.WorkflowStep, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	err = result.All(&steps)
	if err != nil {
		return
	}

	for i := range steps {
		repository.loadMeta(&steps[i])
	}

	return
}

func (repository *repository) FindByID(stepID uint64) (step model.WorkflowStep, err error) {
	result := repository.collection.Find(db.Cond{"id": stepID})
	err = result.One(&step)
	if err != nil {
		return
	}

	err = repository.loadMeta(&step)

	return
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

	return repository.saveMeta(step)
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

	return repository.saveMeta(step)
}

func (repository *repository) loadMeta(step *model.WorkflowStep) (err error) {
	var metas []meta
	result := repository.collectionMeta.Find(db.Cond{"step_id": step.ID})
	if err := result.All(&metas); err != nil {
		return err
	}

	step.Meta = make(map[string]string)
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

func (repository *repository) RemoveByWorkflowID(ID uint64) (err error) {
	var steps []model.WorkflowStep
	res := repository.collection.Find(db.Cond{"workflow_id": ID})
	err = res.All(&steps)
	if err != nil {
		return
	}

	var stepIDs []string // need string for SQL
	for _, s := range steps {
		stepIDs = append(stepIDs, fmt.Sprintf("%d", s.ID))
	}

	// Delete from workflow_steps table
	if err = res.Delete(); err != nil {
		return
	}

	// Delete from workflow_step_meta table
	deleteStepMetaQuery := fmt.Sprintf("DELETE FROM "+workflowStepMetaTableName+" WHERE step_id IN (%s)", strings.Join(stepIDs, ","))
	_, err = repository.collectionMeta.Session().SQL().Exec(deleteStepMetaQuery)

	return
}
