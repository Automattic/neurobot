package matrix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"neurobot/model/message"
	msg "neurobot/model/message"
	"neurobot/model/room"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	mautrixCrypto "maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	mautrixEvent "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
	mautrixId "maunium.net/go/mautrix/id"
)

type mautrixClient interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	JoinRoom(roomIDorAlias, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
	JoinedMembers(roomID id.RoomID) (resp *mautrix.RespJoinedMembers, err error)
	SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID mautrixId.RoomID, eventType mautrixEvent.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	GetPresence(userID id.UserID) (resp *mautrix.RespPresence, err error)
	ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
	SyncWithContext(ctx context.Context) error
}

type mautrixSyncer interface {
	mautrix.Syncer
	mautrix.ExtensibleSyncer
}

type client struct {
	homeserverURL    string
	db               db.Session
	mautrix          mautrixClient
	store            Store // state store
	olmMachine       *mautrixCrypto.OlmMachine
	syncer           mautrixSyncer
	listenersEnabled bool
}

// NewSQLCryptoStore returns a pointer to struct which stores the crypto state
func NewSQLCryptoStore(db db.Session, accountID string, pickleKey []byte) *mautrixCrypto.SQLCryptoStore {
	store := mautrixCrypto.NewSQLCryptoStore(
		db.Driver().(*sql.DB),
		"sqlite3",
		accountID,
		id.DeviceID("NEUROBOT"),
		pickleKey,
		NewApexLogger(),
	)

	if err := store.CreateTables(); err != nil {
		log.WithError(err).WithField("bot", accountID).Fatal("crypto store db migration error")
	}

	return store
}

func DiscoverServerURL(homeserverName string) (homeserverURL *url.URL, err error) {
	var serverURL string
	start := time.Now()

	wellKnown, err := mautrix.DiscoverClientAPI(homeserverName)
	// Both wellKnown and err can be nil for hosts that have https but are not a matrix server.

	if err != nil {
		if strings.Contains(err.Error(), "net/http: TLS handshake timeout") {
			serverURL = "http://" + homeserverName
		} else {
			serverURL = "https://" + homeserverName
		}
	} else {
		if wellKnown != nil {
			serverURL = wellKnown.Homeserver.BaseURL
		} else {
			serverURL = "https://" + homeserverName
		}
	}

	homeserverURL, err = url.Parse(serverURL)
	if err != nil {
		return
	}

	if homeserverURL.Scheme == "" {
		homeserverURL.Scheme = "https"
	}

	if strings.Contains(homeserverURL.Host, "localhost") {
		homeserverURL.Scheme = "http"
	}

	log.WithDuration(time.Since(start)).WithFields(log.Fields{
		"homeserverName": homeserverName,
		"homeserverURL":  homeserverURL.String(),
	}).Info("Discovered homeserver URL")

	return
}

// NewMautrixClient retuns an instance of mautrix client that satisfies Client interface we have defined for a matrix client
func NewMautrixClient(homeserverURL *url.URL, db db.Session, s Store, cryptoStore mautrixCrypto.Store, enableListeners bool) (Client, error) {
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
		Store: s,
	}

	mach := crypto.NewOlmMachine(mautrixClient, NewApexLogger(), cryptoStore, s)
	// Load data from the crypto store
	err := mach.Load()
	if err != nil {
		log.WithError(err).Error("olm machine load error")
		return nil, err
	}

	syncer.OnEventType(mautrixEvent.StateMember, func(source mautrix.EventSource, event *mautrixEvent.Event) {
		mach.HandleMemberEvent(event)
	})
	syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		mach.ProcessSyncResponse(resp, since)
		return true
	})

	client := &client{
		homeserverURL:    homeserverURL.String(),
		db:               db,
		mautrix:          mautrixClient,
		store:            s,
		olmMachine:       mach,
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

func (client *client) OnMessage(handler func(roomID room.ID, message message.Message)) error {
	if err := client.assertListenersEnabled(); err != nil {
		return err
	}

	client.syncer.OnEventType(mautrixEvent.EventMessage, func(source mautrix.EventSource, event *mautrixEvent.Event) {
		var message string
		switch event.Content.AsMessage().MsgType {
		case mautrixEvent.MsgText:
			message = event.Content.AsMessage().Body
		}

		roomID, err := room.NewID(event.RoomID.String())
		if err != nil {
			fmt.Printf("Invalid roomID: %s", err)
			return
		}

		handler(roomID, msg.NewPlainTextMessage(message))
	})

	client.syncer.OnEventType(event.EventEncrypted, func(source mautrix.EventSource, event *mautrixEvent.Event) {
		decrypted, err := client.olmMachine.DecryptMegolmEvent(event)
		if err != nil {
			log.WithError(err).Error("failed to decrypt")
		} else {
			log.WithField("content", decrypted.Content.Raw).Info("received encrypted event")
			message, isMessage := decrypted.Content.Parsed.(*mautrixEvent.MessageEventContent)
			if isMessage && message.Body == "ping" {
				// sendEncrypted(mach, cli, decrypted.RoomID, "Pong!")
			}

			roomID, err := room.NewID(event.RoomID.String())
			if err != nil {
				fmt.Printf("Invalid roomID: %s", err)
				return
			}
			handler(roomID, msg.NewPlainTextMessage(message.Body))
		}
	})

	return nil
}

func (client *client) GetPresence(userID string) string {
	respPresence, err := client.mautrix.GetPresence(id.UserID(userID))
	if err != nil {
		fmt.Printf("error getting presence: %s", err)
		return "unknown"
	}
	return string(respPresence.Presence) // this detection can be improved further if we want to take last_active_ago in account and not rely on presence status alone
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

	// temporarily short-circuit check
	if true || client.store.IsEncrypted(resolvedRoomID) {
		return client.SendEncrypted(roomID, message)
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

func (client *client) SendEncrypted(roomID room.ID, message msg.Message) error {
	resolvedRoomID, err := client.resolveRoomAlias(roomID)
	if err != nil {
		return err
	}

	var content event.MessageEventContent
	switch message.ContentType() {
	case msg.Markdown:
		content = format.RenderMarkdown(message.String(), true, false)
	case msg.PlainText:
		content = event.MessageEventContent{
			MsgType: "m.text",
			Body:    message.String(),
		}
	}

	encrypted, err := client.olmMachine.EncryptMegolmEvent(resolvedRoomID, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = client.olmMachine.ShareGroupSession(resolvedRoomID, client.getRoomMembers(resolvedRoomID))
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"roomID": roomID.ID(),
			}).Error("olm machine error with sharing group session")
			return err
		}
		encrypted, err = client.olmMachine.EncryptMegolmEvent(resolvedRoomID, event.EventMessage, content)
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"roomID": roomID.ID(),
		}).Error("olm machine error with sharing group session")
		return err
	}

	_, err = client.mautrix.SendMessageEvent(resolvedRoomID, event.EventEncrypted, encrypted)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"roomID": roomID.ID(),
		}).Error("error sending encrypted event")
	}

	return err
}

func (client *client) getRoomMembers(roomID mautrixId.RoomID) []mautrixId.UserID {
	var rows []membershipTableRow
	result := client.db.Collection(membershipTable).Find(db.Cond{"room_id": roomID})
	err := result.All(&rows)
	if err != nil {
		if !errors.Is(err, db.ErrNoMoreRows) {
			log.WithError(err).WithField("roomID", roomID).Error("error fetching room members")
		}
		return []id.UserID{}
	}

	users := make([]mautrixId.UserID, len(rows))
	for _, row := range rows {
		users = append(users, mautrixId.UserID(row.UserID))
	}

	// if database has no records for this room, lets fetch and populate it to give an initial state and then events can keep up with membership changes
	if len(users) == 0 {
		resp, err := client.mautrix.JoinedMembers(roomID)
		if err != nil {
			log.WithError(err).WithField("roomID", roomID).Error("error fetching joined members")
			return users
		}

		var row membershipTableRow
		for userID := range resp.Joined {
			row.RoomID = roomID.String()
			row.UserID = userID.String()
			_, err := client.db.Collection(membershipTable).Insert(row)
			if err != nil {
				log.WithError(err).WithField("roomID", roomID).Error("error inserting member info")
			}

			users = append(users, userID)
		}
	}

	return users
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

func (client *client) IsCommand(message msg.Message) bool {
	words := strings.Fields(strings.TrimSpace(message.String()))

	// shortest command can be "!s"
	if len(words) == 0 || len(words[0]) < 2 {
		return false
	}

	if strings.HasPrefix(words[0], "!") {
		return true
	}

	return false
}

func (client *client) assertListenersEnabled() error {
	if !client.listenersEnabled {
		return errors.New("listeners are not enabled. You can enable listeners through the enableListeners argument of NewMautrixClient()")
	}

	return nil
}
