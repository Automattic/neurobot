package matrix

import (
	msg "neurobot/domain/message"
	"neurobot/domain/room"
	"neurobot/mocks"
	"testing"
)

func TestSendPlainTextMessage(t *testing.T) {
	client := NewMautrixClient(mocks.NewMockMatrixClient("bot"))

	roomID, _ := room.NewID("!foo:matrix.test")
	message := msg.NewPlainTextMessage("foo")

	err := client.SendMessage(roomID, message)
	if err != nil {
		t.Error(err)
	}
}

func TestSendMarkdownMessage(t *testing.T) {
	client := NewMautrixClient(mocks.NewMockMatrixClient("bot"))

	roomID, _ := room.NewID("!foo:matrix.test")
	message := msg.NewMarkdownMessage("foo")

	err := client.SendMessage(roomID, message)
	if err != nil {
		t.Error(err)
	}
}
