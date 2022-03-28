package matrix

import (
	msg "neurobot/model/message"
	"neurobot/model/room"
	"neurobot/resources/tests/mocks"
	"testing"
)

func makeClient() (*client, mocks.MautrixClientMock, mocks.MautrixSyncerMock) {
	mautrixMock := mocks.NewMockMatrixClient("bot")
	syncerMock := mocks.NewMockMatrixSyncer()

	client := client{
		homeserverURL: "matrix.test",
		mautrix:       mautrixMock,
		syncer:        syncerMock,
	}

	return &client, mautrixMock, syncerMock
}

func TestSendPlainTextMessage(t *testing.T) {
	client, mautrixMock, _ := makeClient()
	roomID, _ := room.NewID("!foo:matrix.test")
	message := msg.NewPlainTextMessage("foo")

	err := client.SendMessage(roomID, message)
	if err != nil {
		t.Error(err)
	}

	if !mautrixMock.WasMessageSent("foo") {
		t.Error("message: foo wasn't sent")
	}
}

func TestSendMarkdownMessage(t *testing.T) {
	client, mautrixMock, _ := makeClient()
	roomID, _ := room.NewID("!foo:matrix.test")
	message := msg.NewMarkdownMessage("foo")

	err := client.SendMessage(roomID, message)
	if err != nil {
		t.Error(err)
	}

	if !mautrixMock.WasMessageSent("foo") {
		t.Errorf("message: %s wasn't sent", message.String())
	}
}

func TestJoinRoom(t *testing.T) {
	client, mautrixMock, _ := makeClient()
	roomID, _ := room.NewID("!foo:matrix.test")

	err := client.JoinRoom(roomID)
	if err != nil {
		t.Error(err)
	}

	if !mautrixMock.WasRoomJoined("!foo:matrix.test") {
		t.Errorf("room wasn't joined")
	}
}
