package engine

import (
	"errors"
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

type postMessageMatrixWorkflowStepMeta struct {
	messagePrefix string // message prefix
	room          string // Matrix room
	asBot         string // bot identifier, for matrix session
}

type postMessageMatrixWorkflowStep struct {
	workflowStep
	postMessageMatrixWorkflowStepMeta
}

var getMatrixClient = func(homeserver string) (MatrixClient, error) {
	mc, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (s postMessageMatrixWorkflowStep) getMatrixClient(e *engine) (mc MatrixClient, err error) {
	if s.asBot != "" {

		b, err := getBot(e.db, s.asBot)
		if err != nil {
			return nil, err
		}

		if !b.IsHydrated() {
			b.Hydrate(e)
		}

		return b.getMCInstance(), nil
	}

	return e.client, nil
}

func (s postMessageMatrixWorkflowStep) run(p payloadData, e *engine) (payloadData, error) {
	msg := p.Message

	// Append message specified in definition of this step as a prefix to the payload
	if s.messagePrefix != "" {
		if p.Message != "" {
			msg = fmt.Sprintf("%s\n%s", s.messagePrefix, p.Message)
		} else {
			msg = s.messagePrefix
		}
	}

	// Override room defined in meta, if provided in payload
	room := s.room
	if p.Room != "" {
		room = p.Room
	}

	// ensure we have data to work with
	if room == "" {
		return p, errors.New("no room to post")
	}
	if msg == "" {
		return p, errors.New("no message to post")
	}

	mc, err := s.getMatrixClient(e)
	if err != nil {
		return p, err
	}

	// resolve room alias
	if room[0:1] == "#" {
		resolve, err := mc.ResolveAlias(id.RoomAlias(room))
		if err != nil {
			return p, err
		}

		room = resolve.RoomID.String()
	}

	formattedText := format.RenderMarkdown(msg, true, false)
	_, err = mc.SendMessageEvent(id.RoomID(room), event.EventMessage, &formattedText)
	if err != nil {
		return p, err
	}

	return p, nil
}
