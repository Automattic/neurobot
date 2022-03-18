package matrix

import (
	"neurobot/model/message"
	"neurobot/model/room"
)

type Client interface {
	Login(username string, password string) error
	SendMessage(roomID room.ID, message message.Message) error
}
