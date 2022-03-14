package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	mautrixId "maunium.net/go/mautrix/id"
	. "neurobot/domain/message"
	"neurobot/domain/room"
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

func (client *client) SendMessage(roomId room.Id, message Message) error {
	resolvedRoomId, err := client.resolveRoomAlias(roomId)
	if err != nil {
		return err
	}

	switch message.ContentType() {
	case Markdown:
		rendered := format.RenderMarkdown(message.String(), true, false)
		_, err = client.mautrix.SendMessageEvent(resolvedRoomId, event.EventMessage, rendered)

	case PlainText:
		_, err = client.mautrix.SendText(resolvedRoomId, message.String())
	}

	return err
}

func (client *client) resolveRoomAlias(roomId room.Id) (mautrixId.RoomID, error) {
	if !roomId.IsAlias() {
		return mautrixId.RoomID(roomId.Id()), nil
	}

	response, err := client.mautrix.ResolveAlias(mautrixId.RoomAlias(roomId.Id()))
	if err != nil {
		return "", err
	}

	return mautrixId.RoomID(response.RoomID.String()), nil
}
