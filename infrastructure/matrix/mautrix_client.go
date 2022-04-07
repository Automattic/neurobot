package matrix

import (
	"context"
	"errors"
	"fmt"
	"maunium.net/go/mautrix"
	mautrixEvent "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	mautrixId "maunium.net/go/mautrix/id"
	"net/http"
	"net/url"
	msg "neurobot/model/message"
	"neurobot/model/room"
	"strings"
	"time"
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
	mautrix          mautrixClient
	syncer           mautrixSyncer
	listenersEnabled bool
}

func DiscoverServerURL(serverName string) (serverURL string) {
	wellKnown, err := mautrix.DiscoverClientAPI(serverName)
	// Both wellKnown and err can be nil for hosts that have https but are not a matrix server.

	if err != nil {
		if strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
			serverURL = "http://" + serverName
		} else {
			serverURL = "https://" + serverName
		}
	} else {
		if wellKnown != nil {
			serverURL = wellKnown.Homeserver.BaseURL
		} else {
			serverURL = "https://" + serverName
		}
	}

	return serverURL
}

func NewMautrixClient(serverName string, enableListeners bool) (*client, error) {
	homeserverURL, err := url.Parse(DiscoverServerURL(serverName))
	if err != nil {
		return nil, err
	}
	if homeserverURL.Scheme == "" {
		homeserverURL.Scheme = "https"
	}

	var syncer *mautrix.DefaultSyncer
	if enableListeners {
		syncer = mautrix.NewDefaultSyncer()
	} else {
		// TODO: stub syncer?
	}

	mautrixClient := &mautrix.Client{
		AccessToken:   "",
		UserAgent:     mautrix.DefaultUserAgent,
		HomeserverURL: homeserverURL,
		UserID:        "",
		Client:        &http.Client{Timeout: 180 * time.Second},
		Prefix:        mautrix.URLPath{"_matrix", "client", "r0"},
		Syncer:        syncer,
		Logger:        &mautrix.StubLogger{},
		// By default, use an in-memory store which will never save filter ids / next batch tokens to disk.
		// The client will work with this storer: it just won't remember across restarts.
		// In practice, a database backend should be used.
		Store: mautrix.NewInMemoryStore(),
	}

	client := &client{
		homeserverURL:    homeserverURL.String(),
		mautrix:          mautrixClient,
		syncer:           syncer,
		listenersEnabled: enableListeners,
	}

	return client, nil
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
