package engine

import (
	"io/ioutil"
	"log"
	wf "neurobot/app/workflow"
	modelWorkflow "neurobot/model/workflow"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"os"
	"strings"
	"testing"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func TestGetConfiguredWorkflows(t *testing.T) {
	database.Test(func(session db.Session) {
		workflows := fixtures.Workflows(session)
		repository := wf.NewRepository(session)

		got, err := getConfiguredWorkflows(repository)
		if err != nil {
			t.Errorf("could not get configured workflows from database")
		}

		if len(got) != 3 {
			t.Errorf("expected 3 workflows, got %d", len(got))
		}

		var expected []modelWorkflow.Workflow
		expected = append(expected, workflows["QuickStart Demo"])
		expected = append(expected, workflows["MVP"])
		expected = append(expected, workflows["Toml imported Workflow"])

		// have to check just for names, as currently Workflow type in engine is different from workflow model
		// eventually this would be consolidated and a reflect.DeepEqual check should suffice
		for _, w := range expected {
			found := false
			for _, g := range got {
				if w.Name == g.Name {
					found = true
				}
			}
			if !found {
				t.Errorf("output did not match, expected %s in the output", w.Name)
			}
		}

		// if !reflect.DeepEqual(got, expected) {
		// 	t.Errorf("output did not match\n%+v\n%+v", got, expected)
		// }
	})
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
	sqlFiles, err := ioutil.ReadDir("../infrastructure/database/migrations/")
	if err != nil {
		log.Fatal(err)
	}

	var sqls []string
	for _, file := range sqlFiles {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			fileBytes, err := ioutil.ReadFile("../infrastructure/database/migrations/" + file.Name())
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
		// Workflow Steps
		// 'postMatrixMessage' variety (Active)
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (11,'Post message to Matrix room','','postMatrixMessage',11,0,1);`,
		// 'postMatrixMessage' variety (InActive)
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (12,'Deactivated workflow step for matrix room posting','','postMatrixMessage',99,0,0);`,
		// TOML imported workflow's step - 'postMatrixMessage' variety
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (13,'Post message in room 1','description here','postMatrixMessage',13,0,1);`,
		`INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order","active") VALUES (14,'Post message in room 2','description there','postMatrixMessage',13,1,1);`,

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
