package engine

import (
	"testing"
)

func TestGetBot(t *testing.T) {
	dbs, dbs2 := setUp()
	defer tearDown(dbs, dbs2)

	tables := []struct {
		identifier string
		botID      uint64
	}{
		{
			identifier: "bot_afk",
			botID:      2,
		},
		{
			identifier: "bot_none",
			botID:      0, // nil value basically, database row doesn't exist
		},
	}

	for _, table := range tables {
		got, _ := getBot(dbs, table.identifier)
		t.Log(got)
		if table.botID != got.ID {
			t.Errorf("didn't get what was expected. identifier: %s got: %d expected: %d", table.identifier, got.ID, table.botID)
		}
	}
}

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
