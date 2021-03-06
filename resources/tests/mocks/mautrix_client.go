package mocks

import (
	"context"
	"errors"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type MautrixClientMock interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	SendText(roomID id.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	WasMessageSent(text string) bool
	JoinRoom(roomIDorAlias string, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
	WasRoomJoined(roomIDorAlias string) bool
	ResolveAlias(alias id.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
	SyncWithContextWasCalled() bool
	SyncWithContext(ctx context.Context) error
}

type mautrixClientMock struct {
	instantiatedBy        string
	msgs                  []string
	roomsJoined           []string
	syncWithContextCalled bool
}

func NewMautrixClientMock(creator string) MautrixClientMock {
	return &mautrixClientMock{
		instantiatedBy: creator,
	}
}

func (m *mautrixClientMock) Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error) {
	// assume the best for mock purposes
	homeserverInfo := mautrix.HomeserverInfo{BaseURL: ""}
	identityServerInfo := mautrix.IdentityServerInfo{BaseURL: ""}

	return &mautrix.RespLogin{
		AccessToken: "XXXX",
		DeviceID:    "YYYY",
		UserID:      "1",
		WellKnown: &mautrix.ClientWellKnown{
			Homeserver:     homeserverInfo,
			IdentityServer: identityServerInfo,
		},
	}, nil
}

func (m *mautrixClientMock) SendText(roomID id.RoomID, text string) (*mautrix.RespSendEvent, error) {
	// a specific message is designed to return error
	if text == "throwerr" {
		return nil, errors.New("whatever")
	}

	m.msgs = append(m.msgs, text) // store internally for checking, whether this function was called or not

	return &mautrix.RespSendEvent{
		EventID: "AAAA",
	}, nil
}

func (m *mautrixClientMock) SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error) {
	m.msgs = append(m.msgs, contentJSON.(event.MessageEventContent).Body) // store internally for checking, whether this function was called or not

	return &mautrix.RespSendEvent{
		EventID: "AAAA",
	}, nil
}

func (m *mautrixClientMock) WasMessageSent(text string) bool {
	for _, v := range m.msgs {
		if v == text {
			return true
		}
	}

	return false
}

func (m *mautrixClientMock) JoinRoom(roomIDorAlias string, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error) {
	if roomIDorAlias == "" {
		return nil, errors.New("")
	}

	m.roomsJoined = append(m.roomsJoined, roomIDorAlias)

	return
}

func (m *mautrixClientMock) WasRoomJoined(roomIDorAlias string) bool {
	for _, v := range m.roomsJoined {
		if v == roomIDorAlias {
			return true
		}
	}

	return false
}

func (m *mautrixClientMock) ResolveAlias(alias id.RoomAlias) (resp *mautrix.RespAliasResolve, err error) {
	// convert #room:matrix.test to !room:matrix.test as part of mock resolution
	return &mautrix.RespAliasResolve{
		RoomID:  id.RoomID(strings.Replace(alias.String(), "#", "!", 1)),
		Servers: []string{"matrix.test"},
	}, nil
}

func (m *mautrixClientMock) SyncWithContext(ctx context.Context) error {
	m.syncWithContextCalled = true
	return nil
}

func (m *mautrixClientMock) SyncWithContextWasCalled() bool {
	return m.syncWithContextCalled
}
