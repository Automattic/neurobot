package engine

import (
	"fmt"

	"maunium.net/go/mautrix/id"
)

type postMessageMatrixWorkflowStepMeta struct {
	message string
	room    string // Matrix room
}

type postMessageMatrixWorkflowStep struct {
	workflowStep
	postMessageMatrixWorkflowStepMeta
}

func (s postMessageMatrixWorkflowStep) run(payload string, e *engine) (string, error) {
	msg := payload
	// Append message specified in definition of this step as a prefix to the payload
	if s.message != "" {
		if payload != "" {
			msg = fmt.Sprintf("%s %s", s.message, payload)
		} else {
			msg = s.message
		}
	}
	_, err := e.client.SendText(id.RoomID(s.room), msg)
	if err != nil {
		e.log(err.Error())
	}

	return payload, err
}
