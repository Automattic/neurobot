package workflow

import (
	model "neurobot/model/workflow"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"reflect"
	"testing"

	"github.com/upper/db/v4"
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

func TestFindByID(t *testing.T) {
	database.Test(func(session db.Session) {
		workflows := fixtures.Workflows(session)
		repository := NewRepository(session)

		got, err := repository.FindByID(13)
		if err != nil {
			t.Errorf("failed to get workflows by ID: %s", err)
		}

		expected := workflows["Toml imported Workflow"]

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows")
		}
	})
}

func TestFindByIdentifier(t *testing.T) {
	database.Test(func(session db.Session) {
		workflows := fixtures.Workflows(session)
		repository := NewRepository(session)

		got, err := repository.FindByIdentifier("TOMLTEST1")
		if err != nil {
			t.Errorf("failed to get workflows by identifier: %s", err)
		}

		expected := workflows["Toml imported Workflow"]

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active workflows")
		}
	})
}

func TestInsert(t *testing.T) {
	database.Test(func(session db.Session) {
		repository := NewRepository(session)

		workflow := model.Workflow{
			Name:        "foobarbaz-12345",
			Description: "foo",
			Active:      true,
		}

		if err := repository.Save(&workflow); err != nil {
			t.Errorf("failed to insert workflow: %s", err)
		}

		var got model.Workflow
		result := session.Collection("workflows").Find(db.Cond{"id": workflow.ID})
		if err := result.One(&got); err != nil {
			t.Errorf("failed to find workflow: %s", err)
		}

		if !reflect.DeepEqual(got, workflow) {
			t.Errorf("unexpected result insert workflow")
		}
	})
}

func TestUpdate(t *testing.T) {
	database.Test(func(session db.Session) {
		workflows := fixtures.Workflows(session)
		repository := NewRepository(session)

		workflow := workflows["QuickStart Demo"]
		workflow.Name = "updated name"

		if err := repository.Save(&workflow); err != nil {
			t.Errorf("failed to update workflow: %s", err)
		}

		var got model.Workflow
		result := session.Collection("workflows").Find(db.Cond{"id": workflow.ID})
		if err := result.One(&got); err != nil {
			t.Errorf("failed to find workflow: %s", err)
		}

		if got.Name != workflow.Name {
			t.Errorf("failed to update workflow: name was not updated")
		}

		if !reflect.DeepEqual(got, workflow) {
			t.Errorf("unexpected result update workflow")
		}
	})
}

func TestSaveMeta(t *testing.T) {
	database.Test(func(session db.Session) {
		// workflows := fixtures.Workflows(session)
		repository := NewRepository(session)

		w := &model.Workflow{
			Name:        "Toml imported Workflow 3",
			Description: "",
			Active:      true,
			Identifier:  "TOMLTEST3",
		}
		if err := repository.Save(w); err != nil {
			t.Errorf("could not save workflow: %s", err)
		}

		if err := repository.SaveMeta(w); err != nil {
			t.Errorf("could not save workflow meta: %s", err)
		}

		got, err := repository.FindByIdentifier(w.Identifier)
		if err != nil {
			t.Errorf("could not find workflow in the database: %s", err)
		}

		if got.Identifier != w.Identifier {
			t.Errorf("output did not match\n%v\n%v", got, w)
		}
	})
}
