package message

import "strings"

type ContentType int64

const (
	PlainText ContentType = 0
	Markdown              = 1
)

type Message interface {
	ContentType() ContentType
	String() string
	IsCommand() bool
}

type message struct {
	contentType ContentType
	content     string
}

func NewPlainTextMessage(content string) Message {
	return &message{
		contentType: PlainText,
		content:     content,
	}
}

func NewMarkdownMessage(content string) Message {
	return &message{
		contentType: Markdown,
		content:     content,
	}
}

func (message message) ContentType() ContentType {
	return message.contentType
}

func (message message) String() string {
	return message.content
}

func (message message) IsCommand() bool {
	words := strings.Fields(strings.TrimSpace(message.content))

	// shortest command can be "!s"
	if len(words) == 0 || len(words[0]) < 2 {
		return false
	}

	if strings.HasPrefix(words[0], "!") {
		return true
	}

	return false
}
