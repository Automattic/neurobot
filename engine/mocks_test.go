package engine

import (
	"errors"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type mockWorkflowStep struct {
	impact string
}

func (m *mockWorkflowStep) run(payload string, e *engine) (string, error) {
	return payload + m.impact, nil
}

func NewMockWorkflowStep(impact string) *mockWorkflowStep {
	return &mockWorkflowStep{impact: impact}
}

type mockMatrixClient struct {
	msgs []string
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

func NewMockMatrixClient() MatrixClient {
	return &mockMatrixClient{}
}

func NewMockEngine() *engine {
	return &engine{
		client: NewMockMatrixClient(),
	}
}
