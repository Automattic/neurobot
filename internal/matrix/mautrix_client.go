package matrix

import (
	msg "neurobot/domain/message"
	"neurobot/domain/room"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	mautrixId "maunium.net/go/mautrix/id"
)

type mautrixClient interface {
	SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID mautrixId.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
}

type client struct {
	mautrix mautrixClient
}

func NewMautrixClient(mautrix mautrixClient) *client {
	return &client{
		mautrix: mautrix,
	}
}

func (client *client) SendMessage(roomId room.ID, message msg.Message) error {
	resolvedRoomId, err := client.resolveRoomAlias(roomId)
	if err != nil {
		return err
	}

	switch message.ContentType() {
	case msg.Markdown:
		rendered := format.RenderMarkdown(message.String(), true, false)
		_, err = client.mautrix.SendMessageEvent(resolvedRoomId, event.EventMessage, rendered)

	case msg.PlainText:
		_, err = client.mautrix.SendText(resolvedRoomId, message.String())
	}

	return err
}

func (client *client) resolveRoomAlias(roomId room.ID) (mautrixId.RoomID, error) {
	if !roomId.IsAlias() {
		return mautrixId.RoomID(roomId.ID()), nil
	}

	response, err := client.mautrix.ResolveAlias(mautrixId.RoomAlias(roomId.ID()))
	if err != nil {
		return "", err
	}

	return mautrixId.RoomID(response.RoomID.String()), nil
}
