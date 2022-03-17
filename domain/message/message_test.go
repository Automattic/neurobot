package message

import "testing"

func TestNewPlainText(t *testing.T) {
	message := NewPlainTextMessage("foo")

	if message.ContentType() != PlainText {
		t.Error("not a plain text message")
	}
}

func TestNewMarkdown(t *testing.T) {
	message := NewMarkdownMessage("foo")

	if message.ContentType() != Markdown {
		t.Error("not a markdown message")
	}
}

func TestString(t *testing.T) {
	message := NewPlainTextMessage("foo")

	if message.String() != "foo" {
		t.Error("incorrect message content")
	}
}
