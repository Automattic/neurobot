package bot

import (
	"fmt"
	"neurobot/infrastructure/matrix"
	model "neurobot/model/bot"
	"neurobot/model/message"
	"neurobot/model/room"

	"github.com/apex/log"
)

type Registry interface {
	Append(bot model.Bot, client matrix.Client) error
	GetPrimaryClient() (matrix.Client, error)
	GetClient(identifier string) (matrix.Client, error)
}

type registry struct {
	serverName      string
	primaryUsername string
	clients         map[string]matrix.Client
}

func NewRegistry(serverName string) Registry {
	return &registry{
		serverName: serverName,
		clients:    make(map[string]matrix.Client),
	}
}

func (r *registry) Append(bot model.Bot, client matrix.Client) (err error) {
	log.WithFields(log.Fields{"bot": bot.Username}).Info("adding bot to registry")

	if bot.IsPrimary() {
		r.primaryUsername = bot.Username
	}

	if _, ok := r.clients[bot.Username]; ok {
		return fmt.Errorf("bot %s is already known", bot.Username)
	}

	if err = client.Login(bot.Username, bot.Password); err != nil {
		return
	}

	err = client.OnRoomInvite(func(roomID room.ID) {
		// Only accept invitations to rooms in our homeserver
		if roomID.ServerName() != r.serverName {
			fmt.Printf("Ignoring invitation to room in another homeserver: %s", roomID)
			return
		}

		if err := client.JoinRoom(roomID); err != nil {
			fmt.Printf("Failed to join room %s", roomID)
			return
		}
	})

	err = client.OnMessage(func(roomID room.ID, message message.Message) {
		// log messages received temporarily
		log.WithFields(log.Fields{"room": roomID, "bot": bot.Username, "message": message.String()}).Info("message received")
	})

	if err != nil {
		return
	}

	r.clients[bot.Username] = client

	return
}

func (r *registry) GetPrimaryClient() (matrix.Client, error) {
	if client, ok := r.clients[r.primaryUsername]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no primary matrix client was found: %s", r.primaryUsername)
}

func (r *registry) GetClient(identifier string) (matrix.Client, error) {
	if client, ok := r.clients[identifier]; ok {
		return client, nil
	}

	return nil, fmt.Errorf("no matrix client was found for bot with identifier: %s", identifier)
}
