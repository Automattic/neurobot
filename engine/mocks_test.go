package engine

import (
	"errors"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type mockWorkflowPayload struct {
	message string
}

type mockWorkflowStep struct {
	impact string
}

func (m *mockWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	p := payload.(mockWorkflowPayload)

	// a specific payload is designed to return error
	if p.message == "throwerr" {
		return mockWorkflowPayload{message: ""}, errors.New("whatever")
	}

	return mockWorkflowPayload{
		message: p.message + m.impact,
	}, nil
}

func NewMockWorkflowStep(impact string) *mockWorkflowStep {
	return &mockWorkflowStep{impact: impact}
}

type mockMatrixClient struct {
	instantiatedBy string
	msgs           []string
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
