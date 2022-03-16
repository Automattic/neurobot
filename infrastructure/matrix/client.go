package matrix

import (
	"neurobot/domain/message"
	"neurobot/domain/room"
)

type Client interface {
	Login(username string, password string) error
	SendMessage(roomID room.ID, message message.Message) error
}
