package bot

import (
	"github.com/upper/db/v4"
	model "neurobot/model/bot"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"reflect"
	"testing"
)

func TestFindActive(t *testing.T) {
	database.Test(func(session db.Session) {
		bots := fixtures.Bots(session)
		repository := NewRepository(session)

		got, err := repository.FindActive()
		if err != nil {
			t.Errorf("failed to get active bots: %s", err)
		}

		if len(got) < 2 {
			t.Errorf("expected 2 bots, got %d", len(got))
		}

		expected := []model.Bot{bots["active 1"], bots["active 2"]}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected result active bots")
		}
	})
}
