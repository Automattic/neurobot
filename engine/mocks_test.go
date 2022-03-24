package engine

import (
	"errors"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type mockWorkflowStep struct {
	impact string
}

func (m *mockWorkflowStep) run(p map[string]string, e *engine) (map[string]string, error) {
	// a specific payload is designed to return error
	if p["Message"] == "throwerr" {
		return map[string]string{"Message": ""}, errors.New("whatever")
	}

	return map[string]string{
		"Message": p["Message"] + m.impact,
	}, nil
}

func NewMockWorkflowStep(impact string) *mockWorkflowStep {
	return &mockWorkflowStep{impact: impact}
}

type mockMatrixClient struct {
	instantiatedBy string
	msgs           []string
	roomsJoined    []string
}

func (m *mockMatrixClient) Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error) {
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

func (m *mockMatrixClient) SendText(roomID id.RoomID, text string) (*mautrix.RespSendEvent, error) {
	// a specific message is designed to return error
	if text == "throwerr" {
		return nil, errors.New("whatever")
	}

	m.msgs = append(m.msgs, text) // store internally for checking, whether this function was called or not

	return &mautrix.RespSendEvent{
		EventID: "AAAA",
	}, nil
}

func (m *mockMatrixClient) SendMessageEvent(roomID id.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error) {
	m.msgs = append(m.msgs, contentJSON.(*event.MessageEventContent).Body) // store internally for checking, whether this function was called or not

	return &mautrix.RespSendEvent{
		EventID: "AAAA",
	}, nil
}

func (m *mockMatrixClient) WasMessageSent(text string) bool {
	for _, v := range m.msgs {
		if v == text {
			return true
		}
	}

	return false
}

func (m *mockMatrixClient) Sync() error {
	return nil
}

func (m *mockMatrixClient) JoinRoom(roomIDorAlias string, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error) {
	if roomIDorAlias == "" {
		return nil, errors.New("")
	}

	m.roomsJoined = append(m.roomsJoined, roomIDorAlias)

	return
}

func (m *mockMatrixClient) WasRoomJoined(roomIDorAlias string) bool {
	for _, v := range m.roomsJoined {
		if v == roomIDorAlias {
			return true
		}
	}

	return false
}

func (m *mockMatrixClient) ResolveAlias(alias id.RoomAlias) (resp *mautrix.RespAliasResolve, err error) {
	// convert #room:matrix.test to !room:matrix.test as part of mock resolution
	return &mautrix.RespAliasResolve{
		RoomID:  id.RoomID(strings.Replace(alias.String(), "#", "!", 1)),
		Servers: []string{"matrix.test"},
	}, nil
}

func NewMockMatrixClient(creator string) MatrixClient {
	return &mockMatrixClient{
		instantiatedBy: creator,
	}
}

func getMockMatrixClient(homeserver string) (MatrixClient, error) {
	// designed to return error, when invalid url is passed (prefix with space in this case)
	if homeserver != strings.TrimSpace(homeserver) {
		return nil, errors.New("")
	}
	return NewMockMatrixClient("bot"), nil
}

func NewMockEngine() *engine {
	return &engine{
		client: NewMockMatrixClient("engine"),
	}
}
