package matrix

import (
	msg "neurobot/model/message"
	"neurobot/model/room"
	"neurobot/resources/tests/mocks"
	"testing"
)

func makeClient() (*client, mocks.MautrixClientMock) {
	mautrixMock := mocks.NewMockMatrixClient("bot")
	client := NewMautrixClient(mautrixMock)

	return client, mautrixMock
}

func TestSendPlainTextMessage(t *testing.T) {
	client, mautrixMock := makeClient()
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
	client, mautrixMock := makeClient()
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
