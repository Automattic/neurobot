package bot

import (
	model "neurobot/model/bot"

	"github.com/upper/db/v4"
)

type repository struct {
	collection db.Collection
}

func NewRepository(session db.Session) model.Repository {
	return &repository{
		collection: session.Collection("bots"),
	}
}

func (repository *repository) FindActive() (bots []model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"active": 1})
	err = result.All(&bots)
	return
}

func (repository *repository) FindByUsername(username string) (bot model.Bot, err error) {
	result := repository.collection.Find(db.Cond{"username": username})
	err = result.One(&bot)
	return
}

func (repository *repository) Save(bot *model.Bot) (err error) {
	if bot.ID > 0 {
		return repository.update(bot)
	}

	return repository.insert(bot)
}

func (repository *repository) update(bot *model.Bot) (err error) {
	var existing model.Bot

	result := repository.collection.Find(bot.ID)
	if err = result.One(&existing); err != nil {
		return
	}

	existing.Username = bot.Username
	existing.Password = bot.Password
	existing.Description = bot.Description
	existing.Active = bot.Active

	return result.Update(existing)
}

func (repository *repository) insert(bot *model.Bot) (err error) {
	result, err := repository.collection.Insert(bot)
	if err == nil {
		bot.ID = uint64(result.ID().(int64))
	}

	return
}
