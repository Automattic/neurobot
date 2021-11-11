package engine

import (
	"strconv"
	"strings"
	"time"

	"github.com/upper/db/v4"
)

type TriggerRow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Variety     string `db:"variety"`
	Workflows   string `db:"workflow_ids"` // CSV of IDs
	Active      int    `db:"active"`
}
type WorkflowRow struct {
	ID          uint64 `db:"id,omitempty"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Active      int    `db:"active"`
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
					workflows:   splitStringIntoSliceOfInts(row.Workflows, ","),
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
					workflows:   splitStringIntoSliceOfInts(row.Workflows, ","),
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
					message: getWFStepMeta(dbs, row.ID, "message"),
					room:    getWFStepMeta(dbs, row.ID, "room"),
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

	// log.Printf("getTriggerMeta(): id:%d key:%s value:%s\n", trigger_id, key, row["value"])

	return row["value"]
}

func splitStringIntoSliceOfInts(s string, sep string) []uint64 {
	var i []uint64
	for _, piece := range strings.Split(s, sep) {
		if piece == "" {
			continue
		}
		convert, _ := strconv.ParseInt(
			strings.Trim(piece, " "),
			10,
			64,
		)
		i = append(i, uint64(convert))
	}

	return i
}
