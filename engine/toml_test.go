package engine

import (
	"os"
	"reflect"
	"testing"
)

func TestParseTOMLDefs(t *testing.T) {
	e := NewMockEngine()
	tomlFilePath := "./toml_file_for_testing.toml"
	e.workflowsDefTOMLFile = tomlFilePath

	// parseTOMLDefs(e)
	tables := []struct {
		toml    string
		mapping tomlMapping
	}{
		// Empty toml file
		{
			toml:    "",
			mapping: tomlMapping{},
		},
		// Toml file with 2 changed workflows
		// First one, no change
		// Second one, update
		// Third one, insert
		{
			toml: `[[workflow]]
			identifier = "TOMLTEST1"
			active = true
			name = "Toml imported Workflow"

			[[workflow]]
			identifier = "TOMLTEST2"
			active = true # this field is being updated
			name = "Toml imported Workflow 2"

			[[workflow.step]]
			active = true
			variety = "stdout"
			name = "Log to Standard Out"

			[[workflow]]
			identifier = "TOMLTEST3"
			active = true
			name = "Toml imported Workflow 3"`,
			mapping: tomlMapping{
				"TOMLTEST1": 13,
				"TOMLTEST2": 14,
			},
		},
	}

	for _, table := range tables {
		os.WriteFile(tomlFilePath, []byte(table.toml), 0644)

		dbs, dbs2 := setUp()

		preParsingWorkflows, err := getConfiguredWorkflows(dbs)
		if err != nil {
			t.Error("error getting configured workflows")
		}

		e.db = dbs
		parseTOMLDefs(e)

		// check for db changes
		workflows, _ := getConfiguredWorkflows(dbs)
		if table.toml == "" {
			if !reflect.DeepEqual(workflows, preParsingWorkflows) {
				t.Errorf("empty toml file caused a change in workflows in database")
			}
		} else {
			// sufficient to check whether the right operation (insert/update) was trigger on the corresponding workflow
			// actual insert/update functions have their own unit tests

			// last workflow was meant to be inserted
			lastWorkflow := workflows[len(workflows)-1]
			if lastWorkflow.name != "Toml imported Workflow 3" {
				t.Error("toml insert didn't work")
			}

			// second last workflow was meant to be an update
			secondLastWorkflow := workflows[len(workflows)-2]
			// if update (active=true) failed, it won't even be present in the workflows list, as it only picks up active workflows
			if secondLastWorkflow.name != "Toml imported Workflow 2" {
				t.Error("toml update didn't work")
			}
		}

		tearDown(dbs, dbs2)

		os.Remove(tomlFilePath)
	}
}

func TestInsertTOMLWorkflow(t *testing.T) {
	dbs, dbs2 := setUp()

	w := WorkflowTOML{
		Identifier:  "BRANDNEWWORKFLOW",
		Active:      true,
		Name:        "Something to test",
		Description: "Some irregular description",
		Trigger: WorkflowTriggerTOML{
			Name:        "trigger wow",
			Description: "amazing description",
			Variety:     "webhook",
			Meta: map[string]string{
				"urlSuffix": "unittest",
			},
		},
		Steps: []WorkflowStepTOML{
			{
				Active:      true,
				Name:        "Baby Step",
				Description: "Childhood description",
				Variety:     "postMatrixMessage",
				Meta: map[string]string{
					"messagePrefix": "[Alert]",
					"matrixRoom":    "",
				},
			},
			{
				Active:      true,
				Name:        "Another baby Step",
				Description: "Another description",
				Variety:     "postMatrixMessage",
				Meta: map[string]string{
					"messagePrefix": "[Announcement]",
					"matrixRoom":    "",
				},
			},
		},
	}

	insertTOMLWorkflow(dbs, w)

	// examine all these records got inserted at the right places, by directly quering the database
	// check for workflow first
	workflows, err := getConfiguredWorkflows(dbs)
	if err != nil {
		t.Error("couldn't get configured workflows")
	}
	lastWorkflow := workflows[len(workflows)-1]
	if lastWorkflow.name != w.Name || lastWorkflow.description != w.Description {
		t.Errorf("last inserted workflow isn't what we inserted. Got: [%s] Expected: [%s]", lastWorkflow.name, w.Name)
	}

	// check for workflow meta
	got := getWorkflowMeta(dbs, lastWorkflow.id, "toml_identifier")
	if w.Identifier != got {
		t.Errorf("workflow's identifier didn't match. Got: %s Expected: %s", got, w.Identifier)
	}

	// check for trigger
	tr := getWorkflowTrigger(dbs, lastWorkflow.id)
	if tr.Name != w.Trigger.Name || tr.Description != w.Trigger.Description {
		t.Errorf("trigger isn't what we expected. Got: [%s] Expected: [%s]", tr.Name, w.Trigger.Name)
	}

	// check for trigger meta
	for key, expectedValue := range w.Trigger.Meta {
		got := getTriggerMeta(dbs, tr.ID, key)
		if expectedValue != got {
			t.Errorf("trigger meta isn't what's expected. Key:%s Value:%s ExpectedValue:%s", key, got, expectedValue)
		}
	}

	// check for workflow steps
	steps := getWorkflowSteps(dbs, lastWorkflow.id)
	for i, s := range steps {
		if s.Name != w.Steps[i].Name || s.Description != w.Steps[i].Description {
			t.Errorf("workflow step isn't what's expected. Step(%d) Got:%s Expected:%s", i, s.Name, w.Steps[i].Name)
		}

		// check for workflow step meta
		for key, expectedValue := range w.Steps[i].Meta {
			got := getWFStepMeta(dbs, s.ID, key)
			if got != expectedValue {
				t.Errorf("workflow step meta isn't what's expected. Step(%d) Key:%s Value:%s ExpectedValue:%s", i, key, got, expectedValue)
			}
		}
	}

	tearDown(dbs, dbs2)
}

func TestUpdateTOMLWorkflow(t *testing.T) {
	dbs, dbs2 := setUp()

	wid := uint64(13) // TOMLTEST1 identifier workflow is represented by ID 13 in test db
	w := WorkflowTOML{
		Identifier:  "TOMLTEST1",
		Active:      true,
		Name:        "Changed Name",
		Description: "Changed Description",
		Trigger: WorkflowTriggerTOML{
			Name:        "trigger wow",
			Description: "amazing description",
			Variety:     "webhook",
			Meta: map[string]string{
				"urlSuffix": "unittests",
			},
		},
		Steps: []WorkflowStepTOML{
			{
				Active:      true,
				Name:        "Baby Step",
				Description: "Childhood description",
				Variety:     "postMatrixMessage",
				Meta: map[string]string{
					"messagePrefix": "[Alert]",
					"matrixRoom":    "",
				},
			},
			{
				Active:      true,
				Name:        "Another baby Step",
				Description: "Another description",
				Variety:     "postMatrixMessage",
				Meta: map[string]string{
					"messagePrefix": "[Announcement]",
					"matrixRoom":    "",
				},
			},
		},
	}

	err := updateTOMLWorkflow(dbs, wid, w)
	if err != nil {
		t.Errorf("error occured during updation of TOML workflow: %v", err)
	}

	// examine all these records got inserted at the right places, by directly quering the database
	// check for workflow first
	workflows, err := getConfiguredWorkflows(dbs)
	if err != nil {
		t.Error("couldn't get configured workflows")
	}

	found := false

	// find the workflow, because in future, with more data inserts added it could be in between
	for _, workflow := range workflows {

		if w.Identifier != getWorkflowMeta(dbs, workflow.id, "toml_identifier") {
			continue
		}

		found = true

		if workflow.name != w.Name || workflow.description != w.Description {
			t.Errorf("workflow didn't update. Got: [%s] Expected: [%s]", workflow.name, w.Name)
		}

		// check for workflow meta
		got := getWorkflowMeta(dbs, workflow.id, "toml_identifier")
		if w.Identifier != got {
			t.Errorf("workflow's identifier didn't update. Got: %s Expected: %s", got, w.Identifier)
		}

		// check for trigger
		tr := getWorkflowTrigger(dbs, workflow.id)
		if tr.Name != w.Trigger.Name || tr.Description != w.Trigger.Description {
			t.Errorf("trigger didn't update. Got: [%s] Expected: [%s]", tr.Name, w.Trigger.Name)
		}

		// check for trigger meta
		for key, expectedValue := range w.Trigger.Meta {
			got := getTriggerMeta(dbs, tr.ID, key)
			if expectedValue != got {
				t.Errorf("trigger meta didn't update. Key:%s Value:%s ExpectedValue:%s", key, got, expectedValue)
			}
		}

		// check for workflow steps
		steps := getWorkflowSteps(dbs, workflow.id)
		for i, s := range steps {
			if s.Name != w.Steps[i].Name || s.Description != w.Steps[i].Description {
				t.Errorf("workflow step didn't update. Step(%d) Got:%s Expected:%s", i, s.Name, w.Steps[i].Name)
			}

			// check for workflow step meta
			for key, expectedValue := range w.Steps[i].Meta {
				got := getWFStepMeta(dbs, s.ID, key)
				if got != expectedValue {
					t.Errorf("workflow step meta didn't update. Step(%d) Key:%s Value:%s ExpectedValue:%s", i, key, got, expectedValue)
				}
			}
		}
	}

	if !found {
		t.Error("workflow identifier wasn't found in database")
	}

	tearDown(dbs, dbs2)
}

func TestGetTOMLMapping(t *testing.T) {
	dbs, dbs2 := setUp()

	expected := tomlMapping{"TOMLTEST1": 13, "TOMLTEST2": 14}
	got, err := getTOMLMapping(dbs)
	if err != nil {
		t.Error("getTOMLMapping() failed - something wrong with test database")
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("toml mapping mismatch. got: %v expected %v", got, expected)
	}

	_, err = getTOMLMapping(dbs2)
	if err == nil {
		t.Errorf("Error should be thrown, but wasn't.")
	}

	tearDown(dbs, dbs2)
}

func TestRunSemanticCheckOnTOML(t *testing.T) {
	tables := []struct {
		tomlDef   WorkflowDefintionTOML
		shouldErr bool
	}{
		{
			tomlDef:   WorkflowDefintionTOML{},
			shouldErr: false,
		},
		{
			tomlDef: WorkflowDefintionTOML{
				Workflows: []WorkflowTOML{
					{
						Identifier: "Test1",
					},
					{
						Identifier: "Test1",
					},
				},
			},
			shouldErr: true,
		},
		{
			tomlDef: WorkflowDefintionTOML{
				Workflows: []WorkflowTOML{
					{
						Identifier: "Test1",
					},
					{
						Identifier: "Test2",
					},
				},
			},
			shouldErr: false,
		},
	}

	for _, table := range tables {
		err := runSemanticCheckOnTOML(table.tomlDef)
		if err != nil {
			if table.shouldErr == false {
				t.Log(err)
				t.Errorf("runSemanticCheckOnTOML() failed when it shouldn't have")
			}
		} else {
			if table.shouldErr == true {
				t.Log(err)
				t.Errorf("runSemanticCheckOnTOML() didn't fail when it should have")
			}
		}
	}
}

func TestBoolToInt(t *testing.T) {
	tables := []struct {
		input  bool
		output int
	}{
		{
			input:  true,
			output: 1,
		},
		{
			input:  false,
			output: 0,
		},
	}

	for _, table := range tables {
		if table.output != boolToInt(table.input) {
			t.Errorf("boolToInt() failing")
		}
	}
}

func TestAsSha256(t *testing.T) {
	tables := []struct {
		input interface{}
		hash  string
	}{
		{
			input: 1,
			hash:  "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
		},
		{
			// Simple slice of workflow step, variations will follow
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "8bcf9ee916ac1c78319c410fbf8a1b8523ee3f6f613612501651e80073e943c9",
		},
		{
			// Active status is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      false,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "e250badbf5822c4e2bd6cd5a2ed6a0d26e6a2ca79c31fe4c7e6de0802349c2cf",
		},
		{
			// Name is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test My Workflow",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "658e9e3cd92e1c9677327fe60e7dbbef5801662b815261c50882b0eab6480045",
		},
		{
			// Description is changed for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is a different description",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "9af2b40f7be6929f5a6fa2126e52e0508d6f87b30475e2a12a60ec85c0b18805",
		},
		{
			// Different variety of a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is a different description",
						Variety:     "stdout",
						Meta:        map[string]string{},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "d5e01e68bdace2ffa0efc67a19de9baceeef0a5eca9f9cda6777c4dd71763876",
		},
		{
			// Meta value (diff value for a particular meta) for a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "newvalue",
							"key2": "value2",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "56500e217ab3d583400b86304d1f183851c5b855d020d20c0429f2564a874ede",
		},
		{
			// New meta value in a particular workflow step
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key3": "value3",
						},
					},
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "55f1e881fa352e47779b6dc74ed01261bdbf2116eddadaa1daf0d398a8990698",
		},
		{
			// Different count of workflow steps
			input: [][]struct {
				Active      bool
				Name        string
				Description string
				Variety     string
				Meta        map[string]string
			}{
				{
					{
						Active:      true,
						Name:        "Test Workflow Step",
						Description: "This is just to test hashing of workflow steps to identify if they have changed",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"key":  "value",
							"key2": "value2",
						},
					},
				},
			},
			hash: "5b99e2e9b67ce4284628837aa55485ff783356cef5a5e8eb5da0ac6f2f327ae0",
		},
	}

	for _, table := range tables {
		got := asSha256(table.input)
		if got != table.hash {
			t.Errorf("asSha256 hash didn't match. Got: %s Expected: %s", got, table.hash)
		}
	}
}

func TestIntSliceToStringSlice(t *testing.T) {
	tables := []struct {
		intSlice    []uint64
		stringslice []string
	}{
		{
			intSlice:    []uint64{},
			stringslice: []string{},
		},
		{
			intSlice:    []uint64{1},
			stringslice: []string{"1"},
		},
		{
			intSlice:    []uint64{1, 2},
			stringslice: []string{"1", "2"},
		},
		{
			intSlice:    []uint64{1, 2, 3, 4},
			stringslice: []string{"1", "2", "3", "4"},
		},
	}

	for _, table := range tables {
		got := intSliceToStringSlice(table.intSlice)
		if !reflect.DeepEqual(got, table.stringslice) {
			t.Errorf("intSliceToStringSlice didn't generate the correct string slice. Got: %s Expected: %s", got, table.stringslice)
		}
	}
}
