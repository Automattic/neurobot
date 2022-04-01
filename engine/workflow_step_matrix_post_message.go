package engine

import (
	"errors"
	"fmt"
	botApp "neurobot/app/bot"
	"neurobot/infrastructure/matrix"
	"neurobot/model/message"
	r "neurobot/model/room"

	"maunium.net/go/mautrix"
)

type postMessageMatrixWorkflowStepMeta struct {
	messagePrefix string // message prefix
	room          string // Matrix room
	asBot         string // bot identifier, for matrix session
}

type postMessageMatrixWorkflowStep struct {
	workflowStep
	postMessageMatrixWorkflowStepMeta
	botRegistry botApp.Registry
}

var getMatrixClient = func(homeserver string) (MatrixClient, error) {
	mc, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (s postMessageMatrixWorkflowStep) getMatrixClient() (mc matrix.Client, err error) {
	if s.asBot != "" {
		return s.botRegistry.GetPrimaryClient()
	}

	return s.botRegistry.GetClient(s.asBot)
}

func (s postMessageMatrixWorkflowStep) run(p map[string]string) (map[string]string, error) {
	msg := p["Message"]

	// Append message specified in definition of this step as a prefix to the payload
	if s.messagePrefix != "" {
		if p["Message"] != "" {
			msg = fmt.Sprintf("%s\n%s", s.messagePrefix, p["Message"])
		} else {
			msg = s.messagePrefix
		}
	}

	// Override room defined in meta, if provided in payload
	room := s.room
	if p["Room"] != "" {
		room = p["Room"]
	}

	// ensure we have data to work with
	if room == "" {
		return p, errors.New("no room to post")
	}
	if msg == "" {
		return p, errors.New("no message to post")
	}

	mc, err := s.getMatrixClient()
	if err != nil {
		return p, err
	}

	roomID, err := r.NewID(room)
	if err != nil {
		return p, err
	}

	err = mc.SendMessage(roomID, message.NewMarkdownMessage(msg))

	return p, err
}
