package message

type ContentType int64

const (
	PlainText ContentType = 0
	Markdown              = 1
)

type Message interface {
	ContentType() ContentType
	String() string
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
