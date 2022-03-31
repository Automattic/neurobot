package matrix

import (
	"context"
	"errors"
	"fmt"
	"maunium.net/go/mautrix"
	mautrixEvent "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	mautrixId "maunium.net/go/mautrix/id"
	msg "neurobot/model/message"
	"neurobot/model/room"
	"strings"
)

type mautrixClient interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	JoinRoom(roomIDorAlias, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
	SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID mautrixId.RoomID, eventType mautrixEvent.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
	SyncWithContext(ctx context.Context) error
}

type mautrixSyncer interface {
	mautrix.Syncer
	mautrix.ExtensibleSyncer
}

type client struct {
	homeserverURL    string
	homeserverDomain string
	mautrix          mautrixClient
	syncer           mautrixSyncer
	listenersEnabled bool
}

func NewMautrixClient(homeserverURL string, enableListeners bool) (*client, error) {
	mautrixClient, err := mautrix.NewClient(homeserverURL, "", "")
	if err != nil {
		return nil, err
	}

	var syncer mautrixSyncer
	if enableListeners {
		syncer := mautrix.NewDefaultSyncer()
		mautrixClient.Syncer = syncer
	}

	// Remove protocol and port to get just the hostname
	homeserverDomain := strings.Split(homeserverURL, ":")[0]

	client := client{
		homeserverURL:    homeserverURL,
		homeserverDomain: homeserverDomain,
		mautrix:          mautrixClient,
		syncer:           syncer,
		listenersEnabled: enableListeners,
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

	if err != nil {
		return err
	}

	if client.listenersEnabled {
		go func() {
			if err := client.mautrix.SyncWithContext(context.Background()); err != nil {
				fmt.Printf("Error during sync: %s", err)
			}
		}()
	}

	return nil
}

func (client *client) OnRoomInvite(handler func(roomID room.ID)) error {
	if err := client.assertListenersEnabled(); err != nil {
		return err
	}

	client.syncer.OnEventType(mautrixEvent.StateMember, func(source mautrix.EventSource, event *mautrixEvent.Event) {
		if _, ok := event.Content.Raw["membership"]; !ok {
			return
		}

		if event.Content.Raw["membership"] != "invite" {
			return
		}

		roomID, err := room.NewID(event.RoomID.String())
		if err != nil {
			fmt.Printf("Invalid roomID: %s", err)
			return
		}

		handler(roomID)
	})

	return nil
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
		_, err = client.mautrix.SendMessageEvent(resolvedRoomID, mautrixEvent.EventMessage, rendered)

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

func (client *client) assertListenersEnabled() error {
	if !client.listenersEnabled {
		return errors.New("listeners are not enabled. You can enable listeners through the enableListeners argument of NewMautrixClient()")
	}

	return nil
}
