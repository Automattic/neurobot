package engine

import (
	"testing"
)

func TestBotIsHydrated(t *testing.T) {
	b := &Bot{}
	if b.IsHydrated() != false {
		t.Error("empty bot instance should not be hydrated")
	}

	b.hydrated = true
	if b.IsHydrated() != true {
		t.Error("bot instance should be hydrated")
	}

}

func TestBotHydration(t *testing.T) {
	b := &Bot{}
	b.Hydrate(NewMockEngine())

	if b.IsHydrated() != true {
		t.Error("bot instance should have been hydrated")
	}

	if b.e == nil {
		t.Error("bot instance should have reference to engine it was hydrated with")
	}
}

func TestBotGetMCInstance(t *testing.T) {
	b := &Bot{}
	if b.getMCInstance() != nil {
		t.Error("bot matrix instance should have been nil as its not hydrated")
	}

	e := NewMockEngine()
	e.bots = make(map[uint64]MatrixClient)
	e.bots[0] = NewMockMatrixClient("bot")
	b.Hydrate(e)

	if b.getMCInstance() == nil {
		t.Error("bot matrix instance should not have been nil as its now hydrated")
	}
}

func TestBotJoinRoom(t *testing.T) {
	b := &Bot{}
	_, err := b.JoinRoom("whatever")
	if err == nil {
		t.Error("bot shouldn't have been able to join room without hydration")
	}

	e := NewMockEngine()
	e.bots = make(map[uint64]MatrixClient)
	e.bots[0] = NewMockMatrixClient("bot")
	b.Hydrate(e)

	_, err = b.JoinRoom("room1")
	if err != nil {
		t.Error("error thrown while joining a room")
	}

	if !b.getMCInstance().(*mockMatrixClient).WasRoomJoined("room1") {
		t.Error("room wasn't joined when it should have")
	}
}
