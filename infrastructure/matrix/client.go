package matrix

import (
	"neurobot/domain/message"
	"neurobot/domain/room"
)

type Client interface {
	SendMessage(roomId room.Id, message message.Message) error
}
