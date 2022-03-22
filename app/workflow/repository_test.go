package workflow

import (
	"github.com/upper/db/v4"
	model "neurobot/model/workflow"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"reflect"
	"testing"
)

func TestFindActive(t *testing.T) {
	database.Test(func(session db.Session) {
		workflows := fixtures.Workflows(session)
		repository := NewRepository(session)

		got, err := repository.FindActive()
		if err != nil {
			t.Errorf("failed to get active workflows: %s", err)
		}

		if len(got) != 3 {
			t.Errorf("expected 3 workflows, got %d", len(got))
		}

		expected := []model.Workflow{
			workflows["QuickStart Demo"],
			workflows["MVP"],
			workflows["Toml imported Workflow"],
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows")
		}
	})
}
