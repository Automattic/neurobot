package workflowstep

import (
	model "neurobot/model/workflowstep"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"reflect"

	"github.com/upper/db/v4"

	"testing"
)

func TestFindActive(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		repository := NewRepository(session)

		got, err := repository.FindActive()
		if err != nil {
			t.Errorf("failed to get active workflow steps: %s", err)
		}

		if len(got) != 1 {
			t.Errorf("expected 1 workflow steps, got %d", len(got))
		}

		expected := []model.WorkflowStep{
			steps["PostMessage1"],
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows\n%v\n%v", got, expected)
		}
	})
}

func TestFindByID(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		repository := NewRepository(session)

		for _, s := range steps {
			got, err := repository.FindByID(s.ID)
			if err != nil {
				t.Errorf("failed to get workflow step by ID: %s", err)
			}

			expected := s

			if !reflect.DeepEqual(got, expected) {
				t.Errorf("unexpected result\n%v\n%v", got, expected)
			}

		}
	})
}

func TestInsert(t *testing.T) {
	database.Test(func(session db.Session) {
		repository := NewRepository(session)

		step := model.WorkflowStep{
			Name:        "Post to #foo room",
			Description: "foo",
			Variety:     "postMatrixMessage",
			WorkflowID:  1, // irrelevant for this test
			SortOrder:   0,
			Active:      true,
			Meta: map[string]string{
				"messagePrefix": "[Alert]",
				"matrixRoom":    "#foo",
			},
		}

		if err := repository.Save(&step); err != nil {
			t.Errorf("failed to insert workflow step: %s", err)
		}

		got, err := repository.FindByID(step.ID)
		if err != nil {
			t.Errorf("failed to find workflow step: %s", err)
		}

		if !reflect.DeepEqual(got, step) {
			t.Errorf("unexpected result insert workflow step\n%v\n%v", got, step)
		}
	})
}

func TestUpdate(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		repository := NewRepository(session)

		step := steps["PostMessage2"]

		// update fields
		step.Name = "Changed message"
		step.Description = "Changed description"
		step.SortOrder = 2
		step.Active = true
		step.Variety = "bleh"
		step.Meta["matrixRoom"] = "#changed"
		step.Meta["messagePrefix"] = ""

		repository.Save(&step)

		got, err := repository.FindByID(step.ID)
		if err != nil {
			t.Errorf("failed to find workflow step: %s", err)
		}

		if !reflect.DeepEqual(got, step) {
			t.Errorf("unexpected result update workflow step, got: %+v", got)
		}
	})
}
