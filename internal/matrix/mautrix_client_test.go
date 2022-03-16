package matrix

import (
	msg "neurobot/domain/message"
	"neurobot/domain/room"
	"neurobot/mocks"
	"testing"
)

func TestSendPlainTextMessage(t *testing.T) {
	mautrixClient := mocks.NewMockMautrixClient()
	client := NewMautrixClient(mautrixClient)

	roomId, _ := room.NewId("!foo:matrix.test")
	message := msg.NewPlainTextMessage("foo")

	err := client.SendMessage(roomId, message)
	if err != nil {
		t.Error(err)
	}
}

func TestSendMarkdownMessage(t *testing.T) {
	mautrixClient := mocks.NewMockMautrixClient()
	client := NewMautrixClient(mautrixClient)

	roomId, _ := room.NewId("!foo:matrix.test")
	message := msg.NewMarkdownMessage("foo")

	err := client.SendMessage(roomId, message)
	if err != nil {
		t.Error(err)
	}
}
