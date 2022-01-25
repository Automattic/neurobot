package engine

import (
	"log"
	"time"

	"github.com/upper/db/v4"
)

type TriggerRow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Variety     string `db:"variety"`
	WorkflowID  uint64 `db:"workflow_id"`
	Active      int    `db:"active"`
}
type TriggerMetaRow struct {
	ID        uint64 `db:"id,omitempty"`
	TriggerID uint64 `db:"trigger_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}
type WorkflowRow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Active      int    `db:"active"`
}
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

func getConfiguredTriggers(dbs db.Session) (t []Trigger, err error) {
	// get all active triggers out of the database
	var configuredTriggers []TriggerRow
	res := dbs.Collection("triggers").Find(db.Cond{"active": "1"})
	err = res.All(&configuredTriggers)
	if err != nil {
		return
	}

	// range over all active triggers, collecting meta for each trigger and appending that to collect basket
	for _, row := range configuredTriggers {

		switch row.Variety {
		case "webhook":
			t = append(t, &webhookt{
				trigger: trigger{
					id:          row.ID,
					variety:     row.Variety,
					name:        row.Name,
					description: row.Description,
					workflow_id: row.WorkflowID,
				},
				webhooktMeta: webhooktMeta{
					urlSuffix: getTriggerMeta(dbs, row.ID, "urlSuffix"),
				},
			})

		case "poll":
			pollingInterval, _ := time.ParseDuration(getTriggerMeta(dbs, row.ID, "pollingInterval"))
			t = append(t, &pollt{
				trigger: trigger{
					id:          row.ID,
					variety:     row.Variety,
					name:        row.Name,
					description: row.Description,
					workflow_id: row.WorkflowID,
				},
				polltMeta: polltMeta{
					url:             getTriggerMeta(dbs, row.ID, "url"),
					endpointType:    getTriggerMeta(dbs, row.ID, "endpointType"),
					pollingInterval: pollingInterval,
				},
			})
		}
	}

	return
}

func getConfiguredWorkflows(dbs db.Session) (w []workflow, err error) {
	// get all active workflows out of the database
	var savedWorkflows []WorkflowRow
	res := dbs.Collection("workflows").Find(db.Cond{"active": "1"})
	err = res.All(&savedWorkflows)
	if err != nil {
		return
	}

	// range over all active triggers, collecting meta for each trigger and appending that to collect basket
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
	// get all active triggers out of the database
	var configuredSteps []WFStepRow
	res := dbs.Collection("workflow_steps").Find(db.Cond{"active": "1"})
	err = res.All(&configuredSteps)
	if err != nil {
		return
	}

	// range over all active triggers, collecting meta for each trigger and appending that to collect basket
	for _, row := range configuredSteps {
		switch row.Variety {
		case "postMatrixMessage":
			s = append(s, &postMessageMatrixWorkflowStep{
				workflowStep: workflowStep{
					id:          row.ID,
					name:        row.Name,
					description: row.Description,
					variety:     row.Variety,
					workflow_id: row.WorkflowID,
				},
				postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
					messagePrefix: getWFStepMeta(dbs, row.ID, "message"),
					room:          getWFStepMeta(dbs, row.ID, "room"),
				},
			})
		case "stdout":
			s = append(s, &stdoutWorkflowStep{
				workflowStep: workflowStep{
					id:          row.ID,
					name:        row.Name,
					description: row.Description,
					variety:     row.Variety,
					workflow_id: row.WorkflowID,
				},
			})
		}
	}

	return
}

func insertWorkflowMeta(dbs db.Session, id uint64, key string, value string) {
	dbs.Collection("workflow_meta").Insert(WorkflowMetaRow{
		WorkflowID: id,
		Key:        key,
		Value:      value,
	})
}

func insertTriggerMeta(dbs db.Session, id uint64, key string, value string) {
	dbs.Collection("trigger_meta").Insert(TriggerMetaRow{
		TriggerID: id,
		Key:       key,
		Value:     value,
	})
}

func insertWFStepMeta(dbs db.Session, id uint64, key string, value string) {
	dbs.Collection("workflow_step_meta").Insert(WFStepMetaRow{
		StepID: id,
		Key:    key,
		Value:  value,
	})
}

func updateWorkflowMeta(dbs db.Session, workflow_id uint64, key string, value string) {
	res := dbs.Collection("workflow_meta").Find(db.Cond{"workflow_id": workflow_id, "key": key})
	row := make(map[string]string)

	exists, err := res.Exists()
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		insertWorkflowMeta(dbs, workflow_id, key, value)
		return
	}

	res.One(&row)
	if row["value"] == value {
		return
	}

	row["value"] = value
	res.Update(row)
}

func updateTriggerMeta(dbs db.Session, trigger_id uint64, key string, value string) {
	res := dbs.Collection("trigger_meta").Find(db.Cond{"trigger_id": trigger_id, "key": key})
	row := make(map[string]string)

	exists, err := res.Exists()
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		insertTriggerMeta(dbs, trigger_id, key, value)
		return
	}

	res.One(&row)
	if row["value"] == value {
		return
	}

	row["value"] = value
	res.Update(row)
}

func updateWFStepMeta(dbs db.Session, step_id uint64, key string, value string) {
	res := dbs.Collection("workflow_step_meta").Find(db.Cond{"step_id": step_id, "key": key})
	row := make(map[string]string)

	exists, err := res.Exists()
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		insertWFStepMeta(dbs, step_id, key, value)
		return
	}

	res.One(&row)
	if row["value"] == value {
		return
	}

	row["value"] = value
	res.Update(row)
}

func getWorkflowMeta(dbs db.Session, workflow_id uint64, key string) string {
	res := dbs.Collection("workflow_meta").Find(db.Cond{"workflow_id": workflow_id, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getWorkflowMeta(): id:%d key:%s value:%s\n", workflow_id, key, row["value"])

	return row["value"]
}

func getTriggerMeta(dbs db.Session, trigger_id uint64, key string) string {
	res := dbs.Collection("trigger_meta").Find(db.Cond{"trigger_id": trigger_id, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getTriggerMeta(): id:%d key:%s value:%s\n", trigger_id, key, row["value"])

	return row["value"]
}

func getWFStepMeta(dbs db.Session, step_id uint64, key string) string {
	res := dbs.Collection("workflow_step_meta").Find(db.Cond{"step_id": step_id, "key": key})
	row := make(map[string]string)
	res.One(&row)

	// log.Printf("getWFStepMeta(): id:%d key:%s value:%s\n", step_id, key, row["value"])

	return row["value"]
}

func getWorkflowTrigger(dbs db.Session, id uint64) TriggerRow {
	r := TriggerRow{}
	res := dbs.Collection("triggers").Find(db.Cond{"workflow_id": id})
	res.One(&r)
	return r
}
