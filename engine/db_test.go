package engine

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"neurobot/resources/tests/fixtures"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func TestGetConfiguredTriggers(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	var expected []Trigger

	expected = append(expected, Trigger{
		id:          1,
		variety:     "webhook",
		name:        "CURL Request Catcher",
		description: "This webhook trigger will receive your webhook request while showcasing the demo",
		workflowID:  1,
		meta: map[string]string{
			"urlSuffix": "quickstart",
		},
	})
	expected = append(expected, Trigger{
		id:          11,
		variety:     "webhook",
		name:        "Matticspace CURL",
		description: "",
		workflowID:  11,
		meta: map[string]string{
			"urlSuffix": "mcsp",
		},
	})
	expected = append(expected, Trigger{
		id:          14,
		variety:     "webhook",
		name:        "Regular webhook trigger",
		description: "regular description",
		workflowID:  13,
		meta: map[string]string{
			"urlSuffix": "unittest",
		},
	})
	expected = append(expected, Trigger{
		id:          15,
		variety:     "webhook",
		name:        "Regular webhook trigger",
		description: "regular description",
		workflowID:  14,
		meta: map[string]string{
			"urlSuffix": "unittest",
		},
	})

	got, err := getConfiguredTriggers(dbs)
	if err != nil {
		t.Errorf("configured triggers returned an error with database + records")
	} else {
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("configured triggers did not match")
		}
	}

	_, err = getConfiguredTriggers(dbs2)
	if err == nil {
		t.Errorf("configured triggers did not return an error with empty database")
	}
}

func TestGetConfiguredWorkflows(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	var expected []workflow
	expected = append(expected, workflow{
		id:          1,
		name:        "QuickStart Demo",
		description: "This workflow is meant to show a quick demo",
	})
	expected = append(expected, workflow{
		id:          11,
		name:        "MVP",
		description: "",
	})
	expected = append(expected, workflow{
		id:          13,
		name:        "Toml imported Workflow",
		description: "",
	})

	got, err := getConfiguredWorkflows(dbs)
	if err != nil {
		t.Errorf("configured workflows returned an error with database + records")
	} else {
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("configured workflows did not match")
		}
	}

	_, err = getConfiguredWorkflows(dbs2)
	if err == nil {
		t.Errorf("configured workflows did not return an error with empty database")
	}
}

func TestGetConfiguredWFSteps(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	var expected []WorkflowStep
	expected = append(expected, &stdoutWorkflowStep{
		workflowStep: workflowStep{
			id:          1,
			name:        "Log to stdout",
			description: "This workflow step will show the payload to stdout while showcasing the demo",
			variety:     "stdout",
			workflowID:  1,
		},
	})
	expected = append(expected, &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			id:          11,
			name:        "Post message to Matrix room",
			description: "",
			variety:     "postMatrixMessage",
			workflowID:  11,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			messagePrefix: "Alert!",
			room:          "!tnmILBRzpgkBkwSyDY:matrix.test",
		},
	})
	expected = append(expected, &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			id:          13,
			name:        "Post message in room 1",
			description: "description here",
			variety:     "postMatrixMessage",
			workflowID:  13,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			messagePrefix: "[Alert]",
			room:          "",
		},
	})
	expected = append(expected, &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			id:          14,
			name:        "Post message in room 2",
			description: "description there",
			variety:     "postMatrixMessage",
			workflowID:  13,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			messagePrefix: "[Announcement]",
			room:          "",
		},
	})
	//13,'Post message in room 1','description here','postMatrixMessage',13,0,1);`,
	//14,'Post message in room 2','description there','postMatrixMessage',13,1,1);`,

	got, err := getConfiguredWFSteps(dbs)
	if err != nil {
		t.Errorf("configured workflow steps returned an error with database + records")
	} else {
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("configured workflow steps did not match")
			for _, pp := range got {
				t.Log(pp)
			}
		}
	}

	_, err = getConfiguredWFSteps(dbs2)
	if err == nil {
		t.Errorf("configured workflow steps did not return an error with empty database")
	}
}

func TestUpdateWorkflowMeta(t *testing.T) {
	dbs, dbs2 := setUp()

	wid := uint64(11)

	// insert a meta value that doesn't exist, testing insert
	// then update the same meta value, testing update

	key := fmt.Sprintf("neo%d", rand.Intn(100))
	value := "matrix"

	// issue insert
	updateWorkflowMeta(dbs, wid, key, value)
	if value != getWorkflowMeta(dbs, wid, key) {
		t.Error("insert failed")
	}

	value = value + fmt.Sprintf("%d", rand.Intn(100))

	// issue update
	updateWorkflowMeta(dbs, wid, key, value)
	if value != getWorkflowMeta(dbs, wid, key) {
		t.Error("update failed")
	}

	// issue update with same value, which would bail out early (this step slightly increases test coverage)
	updateWorkflowMeta(dbs, wid, key, value)
	if value != getWorkflowMeta(dbs, wid, key) {
		t.Error("update with same value failed")
	}

	// execute once with an empty database to cover returning error for absolute full coverage statistically
	err := updateWorkflowMeta(dbs2, wid, key, value)
	if err == nil {
		t.Error("no error was returned with an empty database with no tables")
	}

	tearDown(dbs, dbs2)
}

func TestUpdateTriggerMeta(t *testing.T) {
	dbs, dbs2 := setUp()

	triggerID := uint64(11)

	// insert a meta value that doesn't exist, testing insert
	// then update the same meta value, testing update

	key := fmt.Sprintf("neo%d", rand.Intn(100))
	value := "matrix"

	// issue insert
	updateTriggerMeta(dbs, triggerID, key, value)
	if value != getTriggerMeta(dbs, triggerID, key) {
		t.Error("insert failed")
	}

	value = value + fmt.Sprintf("%d", rand.Intn(100))

	// issue update
	updateTriggerMeta(dbs, triggerID, key, value)
	if value != getTriggerMeta(dbs, triggerID, key) {
		t.Error("update failed")
	}

	// issue update with same value, which would bail out early (this step slightly increases test coverage)
	updateTriggerMeta(dbs, triggerID, key, value)
	if value != getTriggerMeta(dbs, triggerID, key) {
		t.Error("update with same value failed")
	}

	// execute once with an empty database to cover returning error for absolute full coverage statistically
	err := updateTriggerMeta(dbs2, triggerID, key, value)
	if err == nil {
		t.Error("no error was returned with an empty database with no tables")
	}

	tearDown(dbs, dbs2)
}

func TestUpdateWFStepMeta(t *testing.T) {
	dbs, dbs2 := setUp()

	stepID := uint64(11)

	// insert a meta value that doesn't exist, testing insert
	// then update the same meta value, testing update

	key := fmt.Sprintf("neo%d", rand.Intn(100))
	value := "matrix"

	// issue insert
	updateWFStepMeta(dbs, stepID, key, value)
	if value != getWFStepMeta(dbs, stepID, key) {
		t.Error("insert failed")
	}

	value = value + fmt.Sprintf("%d", rand.Intn(100))

	// issue update
	updateWFStepMeta(dbs, stepID, key, value)
	if value != getWFStepMeta(dbs, stepID, key) {
		t.Error("update failed")
	}

	// issue update with same value, which would bail out early (this step slightly increases test coverage)
	updateWFStepMeta(dbs, stepID, key, value)
	if value != getWFStepMeta(dbs, stepID, key) {
		t.Error("update with same value failed")
	}

	// execute once with an empty database to cover returning error for absolute full coverage statistically
	err := updateWFStepMeta(dbs2, stepID, key, value)
	if err == nil {
		t.Error("no error was returned with an empty database with no tables")
	}

	tearDown(dbs, dbs2)
}

// Function returns two db sessions, first one of a proper database with which tests are meant to pass
// and second one of an empty database with no tables, meant to test errors
func setUp() (db.Session, db.Session) {
	// bump DB log level to fatal errors as triggering an error condition is part of the test
	db.LC().SetLevel(db.LogLevelFatal)

	// Remove sqlite db files, if they exist
	os.Remove("./db_unit_tests.db")
	os.Remove("./db_empty.db")

	// Setup database with some records
	dbs, err := sqlite.Open(sqlite.ConnectionURL{Database: "./db_unit_tests.db"})
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}
	for _, sql := range *getDBSchemaSQL() {
		_, err = dbs.SQL().Exec(sql)
		if err != nil {
			log.Fatal(err)
		}
	}

	fixtures.Bots(dbs)
	fixtures.Workflows(dbs)

	for _, sql := range *getDataInsertsSQL() {
		_, err = dbs.SQL().Exec(sql)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup empty database now
	dbs2, err := sqlite.Open(sqlite.ConnectionURL{Database: "./db_empty.db"})
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	// Return both db sessions
	return dbs, dbs2
}

func tearDown(dbs db.Session, dbs2 db.Session) {
	// Close connections
	dbs.Close()
	dbs2.Close()

	// Delete sqlite db files
	os.Remove("./db_unit_tests.db")
	os.Remove("./db_empty.db")
}

func getDBSchemaSQL() *[]string {
	// read all db schema up files & loop through them to setup the db schema
	sqlFiles, err := ioutil.ReadDir("../infrastructure/database/migration/")
	if err != nil {
		log.Fatal(err)
	}

	var sqls []string
	for _, file := range sqlFiles {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			fileBytes, err := ioutil.ReadFile("../infrastructure/database/migration/" + file.Name())
			if err != nil {
				panic(err)
			}
			sqls = append(sqls, string(fileBytes))
		}
	}

	return &sqls
}

func getDataInsertsSQL() *[]string {
	return &[]string{
		// Triggers
		// 'webhook' variety (Active)
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (11,'Matticspace CURL','','webhook',11,1);`,
		// InActive Trigger (soon to be removed)
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (13,'Disabled Trigger','','webhook',99,0);`,
		// TOML imported workflow's trigger - 'webhook' variety
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (14,'Regular webhook trigger','regular description','webhook',13,1);`,
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (15,'Regular webhook trigger','regular description','webhook',14,1);`,

		// Workflow Steps
		// 'postMatrixMessage' variety (Active)
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (11,'Post message to Matrix room','','postMatrixMessage',11,0,1);`,
		// 'postMatrixMessage' variety (InActive)
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (12,'Deactivated workflow step for matrix room posting','','postMatrixMessage',99,0,0);`,
		// TOML imported workflow's step - 'postMatrixMessage' variety
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (13,'Post message in room 1','description here','postMatrixMessage',13,0,1);`,
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (14,'Post message in room 2','description there','postMatrixMessage',13,1,1);`,

		// Trigger Meta
		// For 'webhook' variety trigger
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (11,11,'urlSuffix','mcsp');`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (15,14,'urlSuffix','unittest');`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (16,15,'urlSuffix','unittest');`,

		// Workflow Step Meta
		// For 'webhook' variety workflow step
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (11,11,'matrixRoom','!tnmILBRzpgkBkwSyDY:matrix.test');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (12,11,'messagePrefix','Alert!');`,
		// TOML imported workflow's step - 'postMatrixMessage' variety
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (13,13,'matrixRoom','');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (14,13,'messagePrefix','[Alert]');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (15,14,'matrixRoom','');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (16,14,'messagePrefix','[Announcement]');`,
	}
}
