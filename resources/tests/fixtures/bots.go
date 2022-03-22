package fixtures

import (
	"github.com/upper/db/v4"
	"neurobot/model/bot"
)

func Bots(session db.Session) map[string]bot.Bot {
	fixtures := map[string]bot.Bot{
		"active 1": {
			ID:          1,
			Name:        "foo",
			Identifier:  "foo",
			Description: "Foo description",
			Username:    "foo_username",
			Password:    "foo_password",
			CreatedBy:   "foo_creator",
			Active:      true,
		},
		"active 2": {
			ID:          2,
			Name:        "bar",
			Identifier:  "bar",
			Description: "Bar description",
			Username:    "bar_username",
			Password:    "bar_password",
			CreatedBy:   "bar_creator",
			Active:      true,
		},
		"inactive": {
			ID:          3,
			Name:        "baz",
			Identifier:  "baz",
			Description: "Baz description",
			Username:    "baz_username",
			Password:    "baz_password",
			CreatedBy:   "baz_creator",
			Active:      false,
		},
	}

	for _, fixture := range fixtures {
		_, _ = session.Collection("bots").Insert(fixture)
	}

	return fixtures
}
