package bot

import (
	"fmt"
	"neurobot/infrastructure/matrix"
	model "neurobot/model/bot"
	"neurobot/model/room"
	"strings"
)

type Registry interface {
	Append(bot model.Bot, client matrix.Client) error
	GetPrimaryClient() (matrix.Client, error)
	GetClient(identifier string) (matrix.Client, error)
}

type registry struct {
	homeserverDomain  string
	primaryIdentifier string
	clients           map[string]matrix.Client
}

func NewRegistry(homeserverURL string) Registry {
	return &registry{
		// Remove port to get just the domain
		homeserverDomain: strings.Split(homeserverURL, ":")[0],
	}
}

func (r *registry) Append(bot model.Bot, client matrix.Client) (err error) {
	if bot.Primary {
		r.primaryIdentifier = bot.Username
	}

	if _, ok := r.clients[bot.Username]; ok {
		return fmt.Errorf("bot %s is already known", bot.Username)
	}

	if err = client.Login(bot.Username, bot.Password); err != nil {
		return
	}

	err = client.OnRoomInvite(func(roomID room.ID) {
		// Only accept invitations to rooms in our homeserver
		if roomID.HomeserverDomain() != r.homeserverDomain {
			fmt.Printf("Ignoring invitation to room in another homeserver: %s", roomID)
			return
		}

		if err := client.JoinRoom(roomID); err != nil {
			fmt.Printf("Failed to join room %s", roomID)
			return
		}
	})

	if err != nil {
		return
	}

	r.clients[bot.Username] = client

	return
}

func (r *registry) GetPrimaryClient() (matrix.Client, error) {
	if client, ok := r.clients[r.primaryIdentifier]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no primary matrix client was found: %s", r.primaryIdentifier)
}

func (r *registry) GetClient(identifier string) (matrix.Client, error) {
	if client, ok := r.clients[identifier]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no matrix client was found for bot with identifier: %s", identifier)
}
