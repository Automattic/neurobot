package mocks

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"time"
)

type MautrixSyncerMock interface {
	mautrix.Syncer
	mautrix.ExtensibleSyncer
}

type mockSyncer struct {
}

func NewMockMatrixSyncer() *mockSyncer {
	return &mockSyncer{}
}

func (mock *mockSyncer) ProcessResponse(resp *mautrix.RespSync, since string) error {
	panic("implement me")
}

func (mock *mockSyncer) OnFailedSync(res *mautrix.RespSync, err error) (time.Duration, error) {
	panic("implement me")
}

func (mock *mockSyncer) GetFilterJSON(userID id.UserID) *mautrix.Filter {
	panic("implement me")
}

func (mock *mockSyncer) OnSync(callback mautrix.SyncHandler) {
	panic("implement me")
}

func (mock *mockSyncer) OnEvent(callback mautrix.EventHandler) {
	panic("implement me")
}

func (mock *mockSyncer) OnEventType(eventType event.Type, callback mautrix.EventHandler) {
	panic("implement me")
}
