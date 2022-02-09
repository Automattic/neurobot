package engine

import (
	"reflect"
	"testing"
)

func TestGetActiveBots(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	expected := []Bot{
		{
			ID:          2,
			Name:        "AFK Bot",
			Identifier:  "bot_afk",
			Description: "Used to post AFK messages for team members",
			Username:    "bot_afk",
			Password:    "bot_afk",
			CreatedBy:   "ashfame",
			Active:      true,
		},
	}

	got, _ := getActiveBots(dbs)
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("incorrect bot data")
	}

	_, err := getActiveBots(dbs2)
	if err == nil {
		t.Error("empty database didn't return error")
	}
}

func TestGetBot(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	tables := []struct {
		identifier string
		botID      uint64
	}{
		{
			identifier: "bot_afk",
			botID:      2,
		},
		{
			identifier: "bot_none",
			botID:      0, // nil value basically, database row doesn't exist
		},
	}

	for _, table := range tables {
		got, _ := getBot(dbs, table.identifier)
		t.Log(got)
		if table.botID != got.ID {
			t.Errorf("didn't get what was expected. identifier: %s got: %d expected: %d", table.identifier, got.ID, table.botID)
		}
	}
}
