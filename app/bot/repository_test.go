package bot

import (
	"github.com/upper/db/v4"
	model "neurobot/model/bot"
	"neurobot/resources/tests/database"
	"neurobot/resources/tests/fixtures"
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	database.Test(func(session db.Session) {
		repository := NewRepository(session)

		bot := model.Bot{
			Username:    "username-12345",
			Password:    "password-12345",
			Description: "foo",
			Active:      true,
			Primary:     false,
		}

		if err := repository.Save(&bot); err != nil {
			t.Errorf("failed to insert bot: %s", err)
		}

		var got model.Bot
		result := session.Collection("bots").Find(db.Cond{"id": bot.ID})
		if err := result.One(&got); err != nil {
			t.Errorf("failed to find bot: %s", err)
		}

		if !reflect.DeepEqual(got, bot) {
			t.Errorf("unexpected result insert bot")
		}
	})
}

func TestUpdate(t *testing.T) {
	database.Test(func(session db.Session) {
		bots := fixtures.Bots(session)
		repository := NewRepository(session)

		bot := bots["active 1"]
		bot.Username = "updated username"
		bot.Password = "updated password"
		bot.Description = "updated description"
		bot.Active = false
		bot.Primary = false

		if err := repository.Save(&bot); err != nil {
			t.Errorf("failed to update bot: %s", err)
		}

		var got model.Bot
		result := session.Collection("bots").Find(db.Cond{"id": bot.ID})
		if err := result.One(&got); err != nil {
			t.Errorf("failed to find bot: %s", err)
		}

		if !reflect.DeepEqual(got, bot) {
			t.Errorf("unexpected result update bot")
		}
	})
}

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

func TestFindByUsername(t *testing.T) {
	database.Test(func(session db.Session) {
		bots := fixtures.Bots(session)
		repository := NewRepository(session)

		bot, err := repository.FindByUsername("bar_username")
		if err != nil {
			t.Errorf("failed to find bot by username: %s", err)
		}

		expected := bots["active 2"]

		if !reflect.DeepEqual(bot, expected) {
			t.Errorf("unexpected result active bots")
		}
	})
}
