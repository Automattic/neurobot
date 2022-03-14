package mocks

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	mautrixId "maunium.net/go/mautrix/id"
)

type MautrixClient interface {
	SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID mautrixId.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
}

type mockMautrixClient struct {
}

func NewMockMautrixClient() MautrixClient {
	return &mockMautrixClient{}
}

func (client *mockMautrixClient) SendText(roomID mautrixId.RoomID, text string) (*mautrix.RespSendEvent, error) {
	return nil, nil
}

func (client *mockMautrixClient) SendMessageEvent(roomID mautrixId.RoomID, eventType event.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error) {
	return nil, nil
}

func (client *mockMautrixClient) ResolveAlias(alias mautrixId.RoomAlias) (resp *mautrix.RespAliasResolve, err error) {
	return nil, nil
}
