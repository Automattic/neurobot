package bot

import (
	"fmt"
	"neurobot/infrastructure/matrix"
	model "neurobot/model/bot"
)

type Registry interface {
	Append(bot model.Bot, client matrix.Client) error
	GetClient(identifier string) matrix.Client
}

type registry struct {
	homeserverURL string
	clients       map[string]matrix.Client
}

func NewRegistry(homeserverURL string) *registry {
	return &registry{homeserverURL: homeserverURL}
}

func (r *registry) Append(bot model.Bot, client matrix.Client) (err error) {
	if _, ok := r.clients[bot.Identifier]; ok {
		return fmt.Errorf("bot %s is already known", bot.Identifier)
	}

	r.clients[bot.Identifier] = client

	return
}

func (r *registry) GetClient(identifier string) matrix.Client {
	if client, ok := r.clients[identifier]; ok {
		return client
	}

	return nil
}
