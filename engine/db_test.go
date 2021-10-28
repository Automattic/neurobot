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
			name:        "Matticspace CURL",
			description: "",
			workflows:   []uint64{1},
		},
		webhooktMeta: webhooktMeta{urlSuffix: "mcsp"},
	})
	expected = append(expected, &pollt{
		trigger: trigger{
			id:          2,
			variety:     "poll",
			name:        "Blog RSS Feed Poller",
			description: "",
			workflows:   []uint64{2},
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
	expected = append(expected, &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			id:          1,
			name:        "Post message to Matrix room",
			description: "",
			variety:     "postMatrixMessage",
			workflow_id: 1,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			message: "Alert!",
			room:    "!tnmILBRzpgkBkwSyDY:matrix.test",
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

func TestSplitStringIntoSliceOfInts(t *testing.T) {
	tables := []struct {
		stringToSplit string
		sep           string
		res           []uint64
	}{
		// test basic functionality
		{
			"100,200",
			",",
			[]uint64{100, 200},
		},
		// test a different separator
		{
			"11-22-33",
			"-",
			[]uint64{11, 22, 33},
		},
		// test with unwanted empty spaces
		{
			" 1 ,2, 3,4 ",
			",",
			[]uint64{1, 2, 3, 4},
		},
		// test invalid input
		{
			"91,92,",
			",",
			[]uint64{91, 92},
		},
	}

	for _, table := range tables {
		got := splitStringIntoSliceOfInts(table.stringToSplit, table.sep)
		if !sliceEquals(table.res, got) {
			t.Errorf("slice of Ints didn't match. got:%v expected:%v", got, table.res)
		}
	}
}

func sliceEquals(a []uint64, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Function returns two db sessions, first one of a proper database with which tests are meant to pass
// and second one of an empty database with no tables, meant to test errors
func setUp() (db.Session, db.Session) {
	// Remove sqlite db files, if they exist
	os.Remove("./db_unit_tests.db")
	os.Remove("./db_empty.db")

	// Setup database with some records
	dbs, err := sqlite.Open(sqlite.ConnectionURL{Database: "./db_unit_tests.db"})
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

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

	for _, sql := range sqls {
		_, err = dbs.SQL().Exec(sql)
		if err != nil {
			log.Fatal(err)
		}
	}

	// insert some test data with which we can test db interaction code
	_, err = dbs.SQL().Exec(`
	INSERT INTO "workflows" ("id","name","description","active") VALUES (1,'MVP','',1);
	INSERT INTO "triggers" ("id","name","description","variety","workflow_ids","active") VALUES (1,'Matticspace CURL','','webhook','1',1);
	INSERT INTO "triggers" ("id","name","description","variety","workflow_ids","active") VALUES (2,'Blog RSS Feed Poller','','poll','2',1);
	INSERT INTO "triggers" ("id","name","description","variety","workflow_ids","active") VALUES (3,'Disabled Trigger','','webhook','99',0);
	INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order") VALUES (1,'Post message to Matrix room','','postMatrixMessage',1,0);
	INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (1,1,'urlSuffix','mcsp');
	INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (2,2,'url','https://wordpress.org/news/feed/');
	INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (3,2,'endpointType','rss');
	INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (4,2,'pollingInterval','1h');
	INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (1,1,'room','!tnmILBRzpgkBkwSyDY:matrix.test');
	INSERT INTO "workflow_step_meta" ("id","step_id","key","value") VALUES (2,1,'message','Alert!');
	`)
	if err != nil {
		log.Fatal(err)
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
