package fixtures

import (
	"github.com/upper/db/v4"
	"log"
	"neurobot/model/bot"
)

func Bots(session db.Session) map[string]bot.Bot {
	fixtures := map[string]bot.Bot{
		"active 1": {
			ID:          1,
			Description: "Foo description",
			Username:    "foo_username",
			Password:    "foo_password",
			Active:      true,
		},
		"active 2": {
			ID:          2,
			Description: "Bar description",
			Username:    "bar_username",
			Password:    "bar_password",
			Active:      true,
		},
		"inactive": {
			ID:          3,
			Description: "Baz description",
			Username:    "baz_username",
			Password:    "baz_password",
			Active:      false,
		},
	}

	for _, fixture := range fixtures {
		_, err := session.Collection("bots").Insert(fixture)
		if err != nil {
			log.Fatalf("Failed to insert fixtures for bots: %s", err)
		}
	}

	return fixtures
}
