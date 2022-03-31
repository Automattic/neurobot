package toml

import (
	"neurobot/app/workflow"
	"neurobot/app/workflowstep"
	wfm "neurobot/model/workflow"
	wfsm "neurobot/model/workflowstep"
	"neurobot/resources/tests/database"
	"os"
	"reflect"
	"testing"

	"github.com/upper/db/v4"
)

func TestImport(t *testing.T) {
	// We only match workflow name to infer Import() worked correctly as Import() don't make workflow IDs for inserted/updated workflows available
	// So we insert a unique name for the workflow and just check for it
	// Edge cases are handled by other unit tests that tests functions used inside of Import() effectively providing good coverage

	// TOML file content in string that we will write to a temporary file
	toml := `[[workflow]]
	identifier = "TOMLTESTME"
	active = true
	name = "Workflow toml test big name workflow"
	description = "some description"

	[[workflow.step]]
	active = true
	name = "Post message"
	description = "Post message to a matrix room"
	variety = "postMatrixMessage"

	[workflow.step.meta]
	messagePrefix = "[Alert]"
	matrixRoom = "#room"`

	tomlFilePath := "./toml_file_for_testing.toml"
	os.WriteFile(tomlFilePath, []byte(toml), 0644)
	defer os.Remove(tomlFilePath)

	database.Test(func(session db.Session) {
		wfRepo := workflow.NewRepository(session)
		wfsRepo := workflowstep.NewRepository(session)

		err := Import(tomlFilePath, wfRepo, wfsRepo)
		if err != nil {
			t.Errorf("valid toml import failed: %s", err)
		}

		expectedWorkflow := wfm.Workflow{Name: "Workflow toml test big name workflow"}

		workflow, err := wfRepo.FindByIdentifier("TOMLTESTME")
		if err != nil {
			t.Errorf("could not find workflow in the database")
		}

		if workflow.Name != expectedWorkflow.Name {
			t.Errorf("imported workflow does not match\n%v\n%v", workflow, expectedWorkflow)
		}
	})
}

func TestParse(t *testing.T) {
	// TOML file content in string that we will write to a temporary file
	toml := `[[workflow]]
	identifier = "TESTME"
	active = true
	name = "Workflow1"
	description = "Some description"

	[[workflow.step]]
	active = true
	name = "Post message"
	description = "Post message to a matrix room"
	variety = "postMatrixMessage"

	[workflow.step.meta]
	messagePrefix = "[Alert]"
	matrixRoom = "#room"`

	tomlFilePath := "./toml_file_for_testing.toml"
	os.WriteFile(tomlFilePath, []byte(toml), 0644)
	defer os.Remove(tomlFilePath)

	got, err := parse(tomlFilePath)
	if err != nil {
		t.Errorf("could not parse toml file: %s", err)
	}

	expected := workflowDefintionTOML{
		Workflows: []workflowTOML{
			{
				Identifier:  "TESTME",
				Active:      true,
				Name:        "Workflow1",
				Description: "Some description",
				Steps: []workflowStepTOML{
					{
						Active:      true,
						Name:        "Post message",
						Description: "Post message to a matrix room",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"messagePrefix": "[Alert]",
							"matrixRoom":    "#room",
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("output did not match\n%v\n%v", got, expected)
	}
}

func TestSemanticCheckOnTOML(t *testing.T) {
	// Testing with valid TOML
	def := workflowDefintionTOML{
		Workflows: []workflowTOML{
			{
				Identifier:  "TESTME",
				Active:      true,
				Name:        "Workflow1",
				Description: "Some description",
				Steps: []workflowStepTOML{
					{
						Active:      true,
						Name:        "Step1",
						Description: "Some description",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"messagePrefix": "[Alert]",
							"matrixRoom":    "",
						},
					},
				},
			},
		},
	}

	err := runSemanticCheckOnTOML(def)
	if err != nil {
		t.Errorf("semantic check on valid toml (def1) failed: %s", err)
	}

	// Testing with invalid TOML - duplicate identifier
	def = workflowDefintionTOML{
		Workflows: []workflowTOML{
			{
				Identifier:  "TESTME",
				Active:      true,
				Name:        "Workflow1",
				Description: "Some description",
				Steps: []workflowStepTOML{
					{
						Active:      true,
						Name:        "Step1",
						Description: "Some description",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"messagePrefix": "[Alert]",
							"matrixRoom":    "",
						},
					},
				},
			},
			{
				Identifier:  "TESTME",
				Active:      true,
				Name:        "Workflow2",
				Description: "Some description",
				Steps: []workflowStepTOML{
					{
						Active:      true,
						Name:        "Step1",
						Description: "Some description",
						Variety:     "postMatrixMessage",
						Meta: map[string]string{
							"messagePrefix": "[Alert]",
							"matrixRoom":    "",
						},
					},
				},
			},
		},
	}

	err = runSemanticCheckOnTOML(def)
	if err == nil {
		t.Errorf("semantic check on invalid toml (duplicate identifier) did not fail")
	}

	// Testing with invalid TOML - no workflow steps defined
	def = workflowDefintionTOML{
		Workflows: []workflowTOML{
			{
				Identifier:  "TESTME",
				Active:      true,
				Name:        "Workflow1",
				Description: "Some description",
			},
		},
	}

	err = runSemanticCheckOnTOML(def)
	if err == nil {
		t.Errorf("semantic check on invalid toml (missing workflow steps) did not fail")
	}
}

func TestPrepare(t *testing.T) {
	database.Test(func(session db.Session) {
		wfRepo := workflow.NewRepository(session)
		wfsRepo := workflowstep.NewRepository(session)

		def := workflowTOML{
			Identifier:  "TESTME",
			Active:      true,
			Name:        "Workflow1",
			Description: "Some description",
			Steps: []workflowStepTOML{
				{
					Active:      true,
					Name:        "Step1",
					Description: "Some description",
					Variety:     "postMatrixMessage",
					Meta: map[string]string{
						"messagePrefix": "[Alert]",
						"matrixRoom":    "",
					},
				},
			},
		}

		preparedWorkflow, preparedSteps, err := prepare(def, wfRepo, wfsRepo)
		if err != nil {
			t.Errorf("could not prepare: %s", err)
		}

		expectedWorkflow := wfm.Workflow{
			Name:        "Workflow1",
			Description: "Some description",
			Active:      true,
			Identifier:  "TESTME",
		}

		expectedSteps := []wfsm.WorkflowStep{
			{
				Name:        "Step1",
				Description: "Some description",
				Variety:     "postMatrixMessage",
				Active:      true,
				Meta: map[string]string{
					"messagePrefix": "[Alert]",
					"matrixRoom":    "",
				},
			},
		}

		if !reflect.DeepEqual(preparedWorkflow, expectedWorkflow) {
			t.Errorf("prepare output not as expected - diff workflow\n%+v\n%+v", preparedWorkflow, expectedWorkflow)
		}
		if !reflect.DeepEqual(preparedSteps, expectedSteps) {
			t.Errorf("prepare output not as expected - diff workflow steps\n%v\n%v", preparedSteps, expectedSteps)
		}
	})
}

// func TestInsertTOMLWorkflow(t *testing.T) {
// 	dbs, dbs2 := setUp()

// 	w := toml.WorkflowTOML{
// 		Identifier:  "BRANDNEWWORKFLOW",
// 		Active:      true,
// 		Name:        "Something to test",
// 		Description: "Some irregular description",
// 		Trigger: toml.WorkflowTriggerTOML{
// 			Name:        "trigger wow",
// 			Description: "amazing description",
// 			Variety:     "webhook",
// 			Meta: map[string]string{
// 				"urlSuffix": "unittest",
// 			},
// 		},
// 		Steps: []toml.WorkflowStepTOML{
// 			{
// 				Active:      true,
// 				Name:        "Baby Step",
// 				Description: "Childhood description",
// 				Variety:     "postMatrixMessage",
// 				Meta: map[string]string{
// 					"messagePrefix": "[Alert]",
// 					"matrixRoom":    "",
// 				},
// 			},
// 			{
// 				Active:      true,
// 				Name:        "Another baby Step",
// 				Description: "Another description",
// 				Variety:     "postMatrixMessage",
// 				Meta: map[string]string{
// 					"messagePrefix": "[Announcement]",
// 					"matrixRoom":    "",
// 				},
// 			},
// 		},
// 	}

// 	insertTOMLWorkflow(dbs, w)

// 	// examine all these records got inserted at the right places, by directly quering the database
// 	// check for workflow first
// 	workflows, err := getConfiguredWorkflows(dbs)
// 	if err != nil {
// 		t.Error("couldn't get configured workflows")
// 	}
// 	lastWorkflow := workflows[len(workflows)-1]
// 	if lastWorkflow.name != w.Name || lastWorkflow.description != w.Description {
// 		t.Errorf("last inserted workflow isn't what we inserted. Got: [%s] Expected: [%s]", lastWorkflow.name, w.Name)
// 	}

// 	// check for workflow meta
// 	got := getWorkflowMeta(dbs, lastWorkflow.id, "toml_identifier")
// 	if w.Identifier != got {
// 		t.Errorf("workflow's identifier didn't match. Got: %s Expected: %s", got, w.Identifier)
// 	}

// 	// check for workflow steps
// 	steps := getWorkflowSteps(dbs, lastWorkflow.id)
// 	for i, s := range steps {
// 		if s.Name != w.Steps[i].Name || s.Description != w.Steps[i].Description {
// 			t.Errorf("workflow step isn't what's expected. Step(%d) Got:%s Expected:%s", i, s.Name, w.Steps[i].Name)
// 		}

// 		// check for workflow step meta
// 		for key, expectedValue := range w.Steps[i].Meta {
// 			got := getWFStepMeta(dbs, s.ID, key)
// 			if got != expectedValue {
// 				t.Errorf("workflow step meta isn't what's expected. Step(%d) Key:%s Value:%s ExpectedValue:%s", i, key, got, expectedValue)
// 			}
// 		}
// 	}

// 	tearDown(dbs, dbs2)
// }

// func TestUpdateTOMLWorkflow(t *testing.T) {
// 	dbs, dbs2 := setUp()

// 	wid := uint64(13) // TOMLTEST1 identifier workflow is represented by ID 13 in test db
// 	w := toml.WorkflowTOML{
// 		Identifier:  "TOMLTEST1",
// 		Active:      true,
// 		Name:        "Changed Name",
// 		Description: "Changed Description",
// 		Trigger: toml.WorkflowTriggerTOML{
// 			Name:        "trigger wow",
// 			Description: "amazing description",
// 			Variety:     "webhook",
// 			Meta: map[string]string{
// 				"urlSuffix": "unittests",
// 			},
// 		},
// 		Steps: []toml.WorkflowStepTOML{
// 			{
// 				Active:      true,
// 				Name:        "Baby Step",
// 				Description: "Childhood description",
// 				Variety:     "postMatrixMessage",
// 				Meta: map[string]string{
// 					"messagePrefix": "[Alert]",
// 					"matrixRoom":    "",
// 				},
// 			},
// 			{
// 				Active:      true,
// 				Name:        "Another baby Step",
// 				Description: "Another description",
// 				Variety:     "postMatrixMessage",
// 				Meta: map[string]string{
// 					"messagePrefix": "[Announcement]",
// 					"matrixRoom":    "",
// 				},
// 			},
// 		},
// 	}

// 	err := updateTOMLWorkflow(dbs, wid, w)
// 	if err != nil {
// 		t.Errorf("error occured during updation of TOML workflow: %v", err)
// 	}

// 	// examine all these records got inserted at the right places, by directly quering the database
// 	// check for workflow first
// 	workflows, err := getConfiguredWorkflows(dbs)
// 	if err != nil {
// 		t.Error("couldn't get configured workflows")
// 	}

// 	found := false

// 	// find the workflow, because in future, with more data inserts added it could be in between
// 	for _, workflow := range workflows {

// 		if w.Identifier != getWorkflowMeta(dbs, workflow.id, "toml_identifier") {
// 			continue
// 		}

// 		found = true

// 		if workflow.name != w.Name || workflow.description != w.Description {
// 			t.Errorf("workflow didn't update. Got: [%s] Expected: [%s]", workflow.name, w.Name)
// 		}

// 		// check for workflow meta
// 		got := getWorkflowMeta(dbs, workflow.id, "toml_identifier")
// 		if w.Identifier != got {
// 			t.Errorf("workflow's identifier didn't update. Got: %s Expected: %s", got, w.Identifier)
// 		}

// 		// check for workflow steps
// 		steps := getWorkflowSteps(dbs, workflow.id)
// 		for i, s := range steps {
// 			if s.Name != w.Steps[i].Name || s.Description != w.Steps[i].Description {
// 				t.Errorf("workflow step didn't update. Step(%d) Got:%s Expected:%s", i, s.Name, w.Steps[i].Name)
// 			}

// 			// check for workflow step meta
// 			for key, expectedValue := range w.Steps[i].Meta {
// 				got := getWFStepMeta(dbs, s.ID, key)
// 				if got != expectedValue {
// 					t.Errorf("workflow step meta didn't update. Step(%d) Key:%s Value:%s ExpectedValue:%s", i, key, got, expectedValue)
// 				}
// 			}
// 		}
// 	}

// 	if !found {
// 		t.Error("workflow identifier wasn't found in database")
// 	}

// 	tearDown(dbs, dbs2)
// }

// func TestGetTOMLMapping(t *testing.T) {
// 	dbs, dbs2 := setUp()

// 	expected := tomlMapping{"TOMLTEST1": 13, "TOMLTEST2": 14}
// 	got, err := getTOMLMapping(dbs)
// 	if err != nil {
// 		t.Error("getTOMLMapping() failed - something wrong with test database")
// 	}

// 	if !reflect.DeepEqual(got, expected) {
// 		t.Errorf("toml mapping mismatch. got: %v expected %v", got, expected)
// 	}

// 	_, err = getTOMLMapping(dbs2)
// 	if err == nil {
// 		t.Errorf("Error should be thrown, but wasn't.")
// 	}

// 	tearDown(dbs, dbs2)
// }

// func TestRunSemanticCheckOnTOML(t *testing.T) {
// 	tables := []struct {
// 		tomlDef   toml.WorkflowDefintionTOML
// 		shouldErr bool
// 	}{
// 		// empty toml is valid toml
// 		{
// 			tomlDef:   toml.WorkflowDefintionTOML{},
// 			shouldErr: false,
// 		},
// 		// no trigger defined
// 		{
// 			tomlDef: toml.WorkflowDefintionTOML{
// 				Workflows: []toml.WorkflowTOML{
// 					{
// 						Identifier: "Test1",
// 						Steps: []toml.WorkflowStepTOML{
// 							{},
// 						},
// 					},
// 				},
// 			},
// 			shouldErr: true,
// 		},
// 		// no workflow steps defined
// 		{
// 			tomlDef: toml.WorkflowDefintionTOML{
// 				Workflows: []toml.WorkflowTOML{
// 					{
// 						Identifier: "Test1",
// 						Trigger: toml.WorkflowTriggerTOML{
// 							Name:        "something",
// 							Description: "something",
// 							Variety:     "webhook",
// 						},
// 					},
// 				},
// 			},
// 			shouldErr: true,
// 		},
// 		// duplicate identifier
// 		{
// 			tomlDef: toml.WorkflowDefintionTOML{
// 				Workflows: []toml.WorkflowTOML{
// 					{
// 						Identifier: "Test1",
// 						Trigger: toml.WorkflowTriggerTOML{
// 							Name:        "something",
// 							Description: "something",
// 							Variety:     "webhook",
// 						},
// 						Steps: []toml.WorkflowStepTOML{
// 							{},
// 						},
// 					},
// 					{
// 						Identifier: "Test1",
// 						Trigger: toml.WorkflowTriggerTOML{
// 							Name:        "something",
// 							Description: "something",
// 							Variety:     "webhook",
// 						},
// 						Steps: []toml.WorkflowStepTOML{
// 							{},
// 						},
// 					},
// 				},
// 			},
// 			shouldErr: true,
// 		},
// 		// valid toml
// 		{
// 			tomlDef: toml.WorkflowDefintionTOML{
// 				Workflows: []toml.WorkflowTOML{
// 					{ture/toml [build f
// 						Identifier: "Test1",
// 						Trigger: toml.WorkflowTriggerTOML{
// 							Name:        "something",
// 							Description: "something",
// 							Variety:     "webhook",
// 						},
// 						Steps: []toml.WorkflowStepTOML{
// 							{},
// 						},
// 					},
// 					{
// 						Identifier: "Test2",
// 						Trigger: toml.WorkflowTriggerTOML{
// 							Name:        "something",
// 							Description: "something",
// 							Variety:     "webhook",
// 						},
// 						Steps: []toml.WorkflowStepTOML{
// 							{},
// 						},
// 					},
// 				},
// 			},
// 			shouldErr: false,
// 		},
// 	}

// 	for _, table := range tables {
// 		err := toml.RunSemanticCheckOnTOML(table.tomlDef)
// 		if err != nil {
// 			if table.shouldErr == false {
// 				t.Log(err)
// 				t.Errorf("runSemanticCheckOnTOML() failed when it shouldn't have")
// 			}
// 		} else {
// 			if table.shouldErr == true {
// 				t.Log(err)
// 				t.Errorf("runSemanticCheckOnTOML() didn't fail when it should have")
// 			}
// 		}
// 	}
// }

// func TestBoolToInt(t *testing.T) {
// 	tables := []struct {
// 		input  bool
// 		output int
// 	}{
// 		{
// 			input:  true,
// 			output: 1,
// 		},
// 		{
// 			input:  false,
// 			output: 0,
// 		},
// 	}

// 	for _, table := range tables {
// 		if table.output != boolToInt(table.input) {
// 			t.Errorf("boolToInt() failing")
// 		}
// 	}
// }

// func TestAsSha256(t *testing.T) {
// 	tables := []struct {
// 		input interface{}
// 		hash  string
// 	}{
// 		{
// 			input: 1,
// 			hash:  "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b",
// 		},
// 		{
// 			// Simple slice of workflow step, variations will follow
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "8bcf9ee916ac1c78319c410fbf8a1b8523ee3f6f613612501651e80073e943c9",
// 		},
// 		{
// 			// Active status is changed for a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      false,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "e250badbf5822c4e2bd6cd5a2ed6a0d26e6a2ca79c31fe4c7e6de0802349c2cf",
// 		},
// 		{
// 			// Name is changed for a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test My Workflow",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "658e9e3cd92e1c9677327fe60e7dbbef5801662b815261c50882b0eab6480045",
// 		},
// 		{
// 			// Description is changed for a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is a different description",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "9af2b40f7be6929f5a6fa2126e52e0508d6f87b30475e2a12a60ec85c0b18805",
// 		},
// 		{
// 			// Different variety of a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is a different description",
// 						Variety:     "stdout",
// 						Meta:        map[string]string{},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "d5e01e68bdace2ffa0efc67a19de9baceeef0a5eca9f9cda6777c4dd71763876",
// 		},
// 		{
// 			// Meta value (diff value for a particular meta) for a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "newvalue",
// 							"key2": "value2",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "56500e217ab3d583400b86304d1f183851c5b855d020d20c0429f2564a874ede",
// 		},
// 		{
// 			// New meta value in a particular workflow step
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key3": "value3",
// 						},
// 					},
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "55f1e881fa352e47779b6dc74ed01261bdbf2116eddadaa1daf0d398a8990698",
// 		},
// 		{
// 			// Different count of workflow steps
// 			input: [][]struct {
// 				Active      bool
// 				Name        string
// 				Description string
// 				Variety     string
// 				Meta        map[string]string
// 			}{
// 				{
// 					{
// 						Active:      true,
// 						Name:        "Test Workflow Step",
// 						Description: "This is just to test hashing of workflow steps to identify if they have changed",
// 						Variety:     "postMatrixMessage",
// 						Meta: map[string]string{
// 							"key":  "value",
// 							"key2": "value2",
// 						},
// 					},
// 				},
// 			},
// 			hash: "5b99e2e9b67ce4284628837aa55485ff783356cef5a5e8eb5da0ac6f2f327ae0",
// 		},
// 	}

// 	for _, table := range tables {
// 		got := asSha256(table.input)
// 		if got != table.hash {
// 			t.Errorf("asSha256 hash didn't match. Got: %s Expected: %s", got, table.hash)
// 		}
// 	}
// }

// func TestIntSliceToStringSlice(t *testing.T) {
// 	tables := []struct {
// 		intSlice    []uint64
// 		stringslice []string
// 	}{
// 		{
// 			intSlice:    []uint64{},
// 			stringslice: []string{},
// 		},
// 		{
// 			intSlice:    []uint64{1},
// 			stringslice: []string{"1"},
// 		},
// 		{
// 			intSlice:    []uint64{1, 2},
// 			stringslice: []string{"1", "2"},
// 		},
// 		{
// 			intSlice:    []uint64{1, 2, 3, 4},
// 			stringslice: []string{"1", "2", "3", "4"},
// 		},
// 	}

// 	for _, table := range tables {
// 		got := intSliceToStringSlice(table.intSlice)
// 		if !reflect.DeepEqual(got, table.stringslice) {
// 			t.Errorf("intSliceToStringSlice didn't generate the correct string slice. Got: %s Expected: %s", got, table.stringslice)
// 		}
// 	}
// }
