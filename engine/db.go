package engine

import (
	"fmt"
	wf "neurobot/app/workflow"
	"neurobot/infrastructure/toml"
	"strings"

	"github.com/upper/db/v4"
)

type WorkflowMetaRow struct {
	ID         uint64 `db:"id,omitempty"`
	WorkflowID uint64 `db:"workflow_id"`
	Key        string `db:"key"`
	Value      string `db:"value"`
}
type WFStepRow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Variety     string `db:"variety"`
	WorkflowID  uint64 `db:"workflow_id"`
	SortOrder   uint64 `db:"sort_order"`
	Active      int    `db:"active"`
}
type WFStepMetaRow struct {
	ID     uint64 `db:"id,omitempty"`
	StepID uint64 `db:"step_id"`
	Key    string `db:"key"`
	Value  string `db:"value"`
}

// get all active workflows out of the database
func getConfiguredWorkflows(dbs db.Session) (w []workflow, err error) {
	repository := wf.NewRepository(dbs)
	savedWorkflows, err := repository.FindActive()
	if err != nil {
		return
	}

	for _, row := range savedWorkflows {
		w = append(w, workflow{
			id:          row.ID,
			name:        row.Name,
			description: row.Description,
		})
	}

	return
}

func getConfiguredWFSteps(dbs db.Session) (s []WorkflowStep, err error) {
	// get all active workflow steps out of the database
	var configuredSteps []WFStepRow
	res := dbs.Collection("workflow_steps").Find(db.Cond{"active": "1"})
	err = res.All(&configuredSteps)
	if err != nil {
		return
	}

	// range over all active steps, collecting meta for each step and appending that to collect basket
	for _, row := range configuredSteps {
		switch row.Variety {
		case "postMatrixMessage":
			s = append(s, &postMessageMatrixWorkflowStep{
				workflowStep: workflowStep{
					id:          row.ID,
					name:        row.Name,
					description: row.Description,
					variety:     row.Variety,
					workflowID:  row.WorkflowID,
				},
				postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
					messagePrefix: getWFStepMeta(dbs, row.ID, "messagePrefix"),
					room:          getWFStepMeta(dbs, row.ID, "matrixRoom"),
					asBot:         getWFStepMeta(dbs, row.ID, "asBot"),
				},
			})
		case "stdout":
			s = append(s, &stdoutWorkflowStep{
				workflowStep: workflowStep{
					id:          row.ID,
					name:        row.Name,
					description: row.Description,
					variety:     row.Variety,
					workflowID:  row.WorkflowID,
				},
			})
		}
	}

	return
}

/**
 * Insert functions for entities (workflow/trigger/step)
 */

func insertWFSteps(dbs db.Session, wid uint64, steps []toml.WorkflowStepTOML) error {
	for i, ws := range steps {
		// insert workflow step
		isr, err := dbs.Collection("workflow_steps").Insert(WFStepRow{
			Name:        ws.Name,
			Description: ws.Description,
			Variety:     ws.Variety,
			WorkflowID:  wid,
			SortOrder:   uint64(i),
			Active:      boolToInt(ws.Active),
		})
		if err != nil {
			return err
		}

		// inserted step ID
		sid := uint64(isr.ID().(int64))

		// insert step meta
		for key, value := range ws.Meta {
			insertWFStepMeta(dbs, sid, key, value)
		}
	}

	return nil
}

/**
 * Insert functions for entities' meta (workflow/trigger/step)
 */

func insertWorkflowMeta(dbs db.Session, id uint64, key string, value string) error {
	_, err := dbs.Collection("workflow_meta").Insert(WorkflowMetaRow{
		WorkflowID: id,
		Key:        key,
		Value:      value,
	})

	return err
}

func insertWFStepMeta(dbs db.Session, id uint64, key string, value string) error {
	_, err := dbs.Collection("workflow_step_meta").Insert(WFStepMetaRow{
		StepID: id,
		Key:    key,
		Value:  value,
	})

	return err
}

/**
 * Update functions for entities' meta (workflow/trigger/step)
 */

func updateWorkflowMeta(dbs db.Session, workflowID uint64, key string, value string) error {
	res := dbs.Collection("workflow_meta").Find(db.Cond{"workflow_id": workflowID, "key": key})
	row := make(map[string]string)

	exists, err := res.Exists()
	if err != nil {
		return err
	}

	if !exists {
		return insertWorkflowMeta(dbs, workflowID, key, value)
	}

	res.One(&row)
	if row["value"] == value {
		return nil
	}

	row["value"] = value
	res.Update(row)

	return nil
}

func updateWFStepMeta(dbs db.Session, stepID uint64, key string, value string) error {
	res := dbs.Collection("workflow_step_meta").Find(db.Cond{"step_id": stepID, "key": key})
	row := make(map[string]string)

	exists, err := res.Exists()
	if err != nil {
		return err
	}

	if !exists {
		insertWFStepMeta(dbs, stepID, key, value)
		return nil
	}

	res.One(&row)
	if row["value"] == value {
		return nil
	}

	row["value"] = value
	res.Update(row)

	return nil
}

/**
 * Get functions for entities' meta (workflow/trigger/step)
 */

func getWorkflowMeta(dbs db.Session, workflowID uint64, key string) string {
	res := dbs.Collection("workflow_meta").Find(db.Cond{"workflow_id": workflowID, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getWorkflowMeta(): id:%d key:%s value:%s\n", workflow_id, key, row["value"])

	return row["value"]
}

func getTriggerMeta(dbs db.Session, triggerID uint64, key string) string {
	res := dbs.Collection("trigger_meta").Find(db.Cond{"trigger_id": triggerID, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getTriggerMeta(): id:%d key:%s value:%s\n", trigger_id, key, row["value"])

	return row["value"]
}

func getWFStepMeta(dbs db.Session, stepID uint64, key string) string {
	res := dbs.Collection("workflow_step_meta").Find(db.Cond{"step_id": stepID, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getWFStepMeta(): id:%d key:%s value:%s\n", step_id, key, row["value"])

	return row["value"]
}

/**
 * Delete functions for entities (workflow/trigger/step)
 */

func deleteAllWFSteps(dbs db.Session, wid uint64) error {
	rows := []WFStepRow{}
	res := dbs.Collection("workflow_steps").Find(db.Cond{"workflow_id": wid})
	res.All(&rows)

	// find all ids for workflow steps, required to delete meta rows
	var collect []uint64
	for _, row := range rows {
		collect = append(collect, row.ID)
	}

	// delete all workfow step rows
	if err := res.Delete(); err != nil {
		return err
	}

	// delete all workflow step meta rows
	return deleteAllWFStepMeta(dbs, collect)
}

func deleteAllWFStepMeta(dbs db.Session, stepIDs []uint64) error {
	_, err := dbs.SQL().Exec(
		fmt.Sprintf(
			"DELETE from workflow_step_meta WHERE step_id IN (%s)",
			strings.Join(
				intSliceToStringSlice(stepIDs),
				",",
			),
		),
	)

	return err
}

func getWorkflowSteps(dbs db.Session, id uint64) []WFStepRow {
	r := []WFStepRow{}
	res := dbs.Collection("workflow_steps").Find(db.Cond{"workflow_id": id})
	res.All(&r)
	return r
}
