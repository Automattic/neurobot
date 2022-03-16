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

func (client *client) SendMessage(roomID room.ID, message msg.Message) error {
	resolvedRoomID, err := client.resolveRoomAlias(roomID)
	if err != nil {
		return err
	}

	switch message.ContentType() {
	case msg.Markdown:
		rendered := format.RenderMarkdown(message.String(), true, false)
		_, err = client.mautrix.SendMessageEvent(resolvedRoomID, event.EventMessage, rendered)

	case msg.PlainText:
		_, err = client.mautrix.SendText(resolvedRoomID, message.String())
	}

	return err
}

func (client *client) resolveRoomAlias(roomID room.ID) (mautrixId.RoomID, error) {
	if !roomID.IsAlias() {
		return mautrixId.RoomID(roomID.ID()), nil
	}

	response, err := client.mautrix.ResolveAlias(mautrixId.RoomAlias(roomID.ID()))
	if err != nil {
		return "", err
	}

	return mautrixId.RoomID(response.RoomID.String()), nil
}
