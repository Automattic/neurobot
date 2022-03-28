package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	mautrixId "maunium.net/go/mautrix/id"
	msg "neurobot/model/message"
	"neurobot/model/room"
)

type mautrixClient interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	JoinRoom(roomIDorAlias, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
	SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID mautrixId.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
}

type client struct {
	homeserverURL string
	mautrix       mautrixClient
}

func NewMautrixClient(homeserverURL string) (*client, error) {
	mautrixClient, err := mautrix.NewClient(homeserverURL, "", "")
	if err != nil {
		return nil, err
	}

	client := client{
		homeserverURL: homeserverURL,
		mautrix:       mautrixClient,
	}

	return &client, nil
}

func (client *client) Login(username string, password string) error {
	_, err := client.mautrix.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: username},
		Password:         password,
		DeviceID:         "NEUROBOT",
		StoreCredentials: true,
	})

	return err
}

func (client *client) JoinRoom(id room.ID) (err error) {
	_, err = client.mautrix.JoinRoom(id.ID(), "", "")

	return
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
