package matrix

import (
	"neurobot/model/message"
	msg "neurobot/model/message"
	"neurobot/model/room"
	"neurobot/resources/tests/mocks"
	"testing"
)

func makeClient() (*client, mocks.MautrixClientMock, mocks.MautrixSyncerMock) {
	mautrixMock := mocks.NewMautrixClientMock("bot")
	syncerMock := mocks.NewMockMatrixSyncer()

	client := client{
		homeserverURL:    "matrix.test",
		mautrix:          mautrixMock,
		syncer:           syncerMock,
		listenersEnabled: false,
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

func TestIsCommand(t *testing.T) {
	client, _, _ := makeClient()

	tables := []struct {
		description string
		msg         string
		isCommand   bool
	}{
		{
			description: "empty message",
			msg:         "",
			isCommand:   false,
		},
		{
			description: "regular message",
			msg:         "hello john",
			isCommand:   false,
		},
		{
			description: "command with no arguments",
			msg:         "!echo",
			isCommand:   true,
		},
		{
			description: "command with 1 argument",
			msg:         "!echo sound",
			isCommand:   true,
		},
		{
			description: "command with arguments",
			msg:         "!echo sound everywhere",
			isCommand:   true,
		},
		{
			description: "broken command - missing command name",
			msg:         "!",
			isCommand:   false,
		},
		{
			description: "broken command - missing command name but with arguments",
			msg:         "! sound everywhere",
			isCommand:   false,
		},
	}

	for _, table := range tables {
		output := client.IsCommand(message.NewPlainTextMessage(table.msg))
		if table.isCommand != output {
			t.Errorf("mismatch for [%s] should be %t but is %t", table.description, table.isCommand, output)
		}
	}

}
