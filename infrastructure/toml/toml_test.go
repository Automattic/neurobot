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
