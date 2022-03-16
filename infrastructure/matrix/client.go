package matrix

import (
	"neurobot/domain/message"
	"neurobot/domain/room"
)

type Client interface {
	SendMessage(roomID room.ID, message message.Message) error
}
