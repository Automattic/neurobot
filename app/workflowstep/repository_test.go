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

func TestFindActiveSortOrder(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		repository := NewRepository(session)

		// Test for sort_order now
		expected := []model.WorkflowStep{
			steps["PostMessage3"],
			steps["PostMessage1"],
		}

		// Setup needed db state
		// enable PostMessage3 step
		step, err := repository.FindByID(3)
		if err != nil {
			t.Errorf("error setting up state for sort_order test: %s", err)
		}
		step.Active = true
		if err := repository.Save(&step); err != nil {
			t.Errorf("error setting up state for sort_order test: %s", err)
		}
		expected[0] = step

		// change PostMessage1 sort order
		step, err = repository.FindByID(1)
		if err != nil {
			t.Errorf("error setting up state for sort_order test: %s", err)
		}
		step.SortOrder = 2
		if err := repository.Save(&step); err != nil {
			t.Errorf("error setting up state for sort_order test: %s", err)
		}
		expected[1] = step

		got, err := repository.FindActive()
		if err != nil {
			t.Errorf("failed to get active workflow steps: %s", err)
		}

		if len(got) != 2 {
			t.Errorf("expected 2 workflow steps, got %d", len(got))
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows (sort order)\n%+v\n%+v", got, expected)
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

func TestFindByWorkflowID(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		repository := NewRepository(session)

		got, err := repository.FindByWorkflowID(1)
		if err != nil {
			t.Errorf("failed to get workflow step: %s", err)
		}

		expected := []model.WorkflowStep{
			steps["PostMessage1"],
			steps["PostMessage3"],
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows\n%v\n%v", got, expected)
		}
	})
}

func TestRemoveByWorkflowID(t *testing.T) {
	database.Test(func(session db.Session) {
		steps := fixtures.WorkflowSteps(session)
		stepToBeDeleted := steps["PostMessage1"]

		// t.Errorf("%+v", steps)
		repository := NewRepository(session)

		if err := repository.RemoveByWorkflowID(stepToBeDeleted.ID); err != nil {
			t.Errorf("unable to remove workflow steps based on workflow ID")
		}

		got, err := repository.FindActive()
		if err != nil {
			t.Errorf("error querying for workflow steps")
		}

		if len(got) == 1 {
			t.Errorf("workflow steps were not deleted")
		}

		gotCount, err := session.Collection(workflowStepMetaTableName).Find(db.Cond{"step_id": stepToBeDeleted.ID}).Count()
		if err != nil {
			t.Errorf("could not get data out of workflow step meta table")
		}

		if gotCount > 0 {
			t.Errorf("workflow step meta was not deleted")
		}
	})
}
