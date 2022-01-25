package engine

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func TestGetConfiguredTriggers(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	var expected []Trigger
	expected = append(expected, &webhookt{
		trigger: trigger{
			id:          1,
			variety:     "webhook",
			name:        "CURL Request Catcher",
			description: "This webhook trigger will receive your webhook request while showcasing the demo",
			workflow_id: 1,
		},
		webhooktMeta: webhooktMeta{urlSuffix: "quickstart"},
	})
	expected = append(expected, &webhookt{
		trigger: trigger{
			id:          11,
			variety:     "webhook",
			name:        "Matticspace CURL",
			description: "",
			workflow_id: 11,
		},
		webhooktMeta: webhooktMeta{urlSuffix: "mcsp"},
	})
	expected = append(expected, &pollt{
		trigger: trigger{
			id:          12,
			variety:     "poll",
			name:        "Blog RSS Feed Poller",
			description: "",
			workflow_id: 12,
		},
		polltMeta: polltMeta{
			url:             "https://wordpress.org/news/feed/",
			endpointType:    "rss",
			pollingInterval: time.Hour,
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
			workflow_id: 1,
		},
	})
	expected = append(expected, &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			id:          11,
			name:        "Post message to Matrix room",
			description: "",
			variety:     "postMatrixMessage",
			workflow_id: 11,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			messagePrefix: "Alert!",
			room:          "!tnmILBRzpgkBkwSyDY:matrix.test",
		},
	})

	got, err := getConfiguredWFSteps(dbs)
	if err != nil {
		t.Errorf("configured workflow steps returned an error with database + records")
	} else {
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("configured workflow steps did not match")
		}
	}

	_, err = getConfiguredWFSteps(dbs2)
	if err == nil {
		t.Errorf("configured workflow steps did not return an error with empty database")
	}
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
	sqlFiles, err := ioutil.ReadDir("../migration/")
	if err != nil {
		log.Fatal(err)
	}

	var sqls []string
	for _, file := range sqlFiles {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			fileBytes, err := ioutil.ReadFile("../migration/" + file.Name())
			if err != nil {
				panic(err)
			}
			sqls = append(sqls, string(fileBytes))
		}
	}

	return &sqls
}

func getDataInsertsSQL() *[]string {
	// @TODO add comments and more entries to better cover different set of possibilities
	return &[]string{
		`INSERT INTO "workflows" ("id","name","description","active") VALUES (11,'MVP','',1);`,
		`INSERT INTO "workflows" ("id","name","description","active") VALUES (12,'Deactivated Workflow','',0);`,
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (11,'Matticspace CURL','','webhook','11',1);`,
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (12,'Blog RSS Feed Poller','','poll','12',1);`,
		`INSERT INTO "triggers" ("id","name","description","variety","workflow_id","active") VALUES (13,'Disabled Trigger','','webhook','99',0);`,
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (11,'Post message to Matrix room','','postMatrixMessage',11,0,1);`,
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (12,'Deactivated workflow step for matrix room posting','','postMatrixMessage',99,0,0);`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (11,11,'urlSuffix','mcsp');`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (12,12,'url','https://wordpress.org/news/feed/');`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (13,12,'endpointType','rss');`,
		`INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (14,12,'pollingInterval','1h');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (11,11,'room','!tnmILBRzpgkBkwSyDY:matrix.test');`,
		`INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (12,11,'message','Alert!');`,
	}
}
