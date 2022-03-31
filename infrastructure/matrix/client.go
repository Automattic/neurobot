package matrix

import (
	"neurobot/model/message"
	"neurobot/model/room"
)

type Client interface {
	Login(username string, password string) error
	JoinRoom(id room.ID) error
	SendMessage(roomID room.ID, message message.Message) error

	// OnRoomInvite registers a handler that will be called whenever the currently authenticated user is invited to a room.
	OnRoomInvite(handler func(roomID room.ID)) error
}
