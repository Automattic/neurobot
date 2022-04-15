package matrix

import (
	"neurobot/resources/tests/database"
	"reflect"
	"testing"

	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func TestSaveFilterID(t *testing.T) {
	database.Test(func(session db.Session) {
		botID := uint64(1)
		storer := NewStorer(session, botID)

		scenarios := []struct {
			id  string
			val string
		}{
			{
				id:  "foo",
				val: "xxxx",
			},
			{
				id:  "bar",
				val: "",
			},
		}

		for _, scenario := range scenarios {
			storer.SaveFilterID(id.UserID(scenario.id), scenario.val)

			var got row
			result := session.Collection(table).Find(db.Cond{
				"bot_id": botID,
				"what":   "filter",
				"id":     scenario.id,
			})
			if err := result.One(&got); err != nil {
				t.Errorf("failed to read from database: %s", err)
			}

			if scenario.val != got.Value {
				t.Errorf("did not return expected value, got: %+v expected: %+v", got, scenario.val)
			}
		}
	})
}

func TestLoadFilterID(t *testing.T) {
	database.Test(func(session db.Session) {
		botID := uint64(1)
		storer := NewStorer(session, botID)

		scenarios := []struct {
			id  string
			val string
		}{
			{
				id:  "foo",
				val: "yyyy",
			},
			{
				id:  "bar",
				val: "",
			},
		}

		for _, scenario := range scenarios {
			if _, err := session.Collection(table).Insert(row{
				BotID: botID,
				What:  "filter",
				ID:    scenario.id,
				Value: scenario.val,
			}); err != nil {
				t.Errorf("error inserting data for read: %s", err)
			}

			got := storer.LoadFilterID(id.UserID(scenario.id))
			if scenario.val != got {
				t.Errorf("did not get expected value, got: %s", got)
			}
		}
	})
}

func TestSaveNextBatch(t *testing.T) {
	database.Test(func(session db.Session) {
		botID := uint64(1)
		storer := NewStorer(session, botID)

		scenarios := []struct {
			id  string
			val string
		}{
			{
				id:  "foo",
				val: "xxxx",
			},
			{
				id:  "bar",
				val: "",
			},
		}

		for _, scenario := range scenarios {
			storer.SaveNextBatch(id.UserID(scenario.id), scenario.val)

			var got row
			result := session.Collection(table).Find(db.Cond{
				"bot_id": botID,
				"what":   "batch",
				"id":     scenario.id,
			})
			if err := result.One(&got); err != nil {
				t.Errorf("failed to read from database: %s", err)
			}

			if scenario.val != got.Value {
				t.Errorf("did not return expected value, got: %+v expected: %+v", got, scenario.val)
			}
		}
	})
}

func TestLoadNextBatch(t *testing.T) {
	database.Test(func(session db.Session) {
		botID := uint64(1)
		storer := NewStorer(session, botID)

		scenarios := []struct {
			id  string
			val string
		}{
			{
				id:  "foo",
				val: "yyyy",
			},
			{
				id:  "bar",
				val: "",
			},
		}

		for _, scenario := range scenarios {
			if _, err := session.Collection(table).Insert(row{
				BotID: botID,
				What:  "batch",
				ID:    scenario.id,
				Value: scenario.val,
			}); err != nil {
				t.Errorf("error inserting data for read: %s", err)
			}

			got := storer.LoadNextBatch(id.UserID(scenario.id))
			if scenario.val != got {
				t.Errorf("did not get expected value, got: %s", got)
			}
		}
	})
}

func TestSaveAndLoadRoom(t *testing.T) {
	database.Test(func(session db.Session) {
		botID := uint64(1)
		storer := NewStorer(session, botID)

		var event event.Event
		roomID := id.RoomID("!xxxxxxx@example.org")

		eventJSON := `
		{
			"content": {
			  "membership": "join"
			},
			"event_id": "$26RqwJMLw-yds1GAH_QxjHRC1Da9oasK0e5VLnck_45",
			"origin_server_ts": 1632489532305,
			"room_id": "!jEsUZKDJdhlrceRyVU:example.org",
			"sender": "@example:example.org",
			"state_key": "@user:example.org",
			"type": "m.room.member",
			"unsigned": {
			  "age": 1567437,
			  "redacted_because": {
				"content": {
				  "reason": "spam"
				},
				"event_id": "$Nhl3rsgHMjk-DjMJANawr9HHAhLg4GcoTYrSiYYGqEE",
				"origin_server_ts": 1632491098485,
				"redacts": "$26RqwJMLw-yds1GAH_QxjHRC1Da9oasK0e5VLnck_45",
				"room_id": "!jEsUZKDJdhlrceRyVU:example.org",
				"sender": "@moderator:example.org",
				"type": "m.room.redaction",
				"unsigned": {
				  "age": 1257
				}
			  }
			}
		  }
		`
		if err := event.UnmarshalJSON([]byte(eventJSON)); err != nil {
			t.Errorf("error while unmarshaling sample JSON for event: %s", err)
		}

		room := mautrix.NewRoom(roomID)
		room.UpdateState(&event)

		storer.SaveRoom(room)
		gotRoom := storer.LoadRoom(roomID)

		if !reflect.DeepEqual(room, gotRoom) {
			t.Errorf("rooms did not match\ngot: %+v\nexpected: %+v", gotRoom, room)
		}
	})
}
