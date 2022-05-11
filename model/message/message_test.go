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

func TestIsCommand(t *testing.T) {
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
		output := NewPlainTextMessage(table.msg).IsCommand()
		if table.isCommand != output {
			t.Errorf("mismatch for [%s] should be %t but is %t", table.description, table.isCommand, output)
		}
	}
}
