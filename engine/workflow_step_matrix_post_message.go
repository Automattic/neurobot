package engine

import (
	"errors"
	"fmt"

	"maunium.net/go/mautrix"
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

func (s postMessageMatrixWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	if payload == nil {
		// nothing to do, let the next workflow step continue
		return nil, nil
	}
	p := payload.(payloadData)
	msg := p.Message

	// Append message specified in definition of this step as a prefix to the payload
	if s.messagePrefix != "" {
		if p.Message != "" {
			msg = fmt.Sprintf("%s\n%s", s.messagePrefix, p.Message)
		} else {
			msg = s.messagePrefix
		}
	}

	mc, err := s.getMatrixClient(e)
	if err != nil {
		return nil, err
	}

	// Override room defined in meta, if provided in payload
	room := s.room
	if p.Room != "" {
		room = p.Room
	}

	if room == "" {
		return nil, errors.New("no room to post")
	}

	_, err = mc.SendText(id.RoomID(room), msg)
	if err != nil {
		return payload, err
	}

	return payload, nil
}
