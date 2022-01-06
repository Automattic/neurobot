package engine

import (
	"fmt"

	"maunium.net/go/mautrix/id"
)

type postMessageMatrixWorkflowPayload struct {
	message string
	room    string
}

type postMessageMatrixWorkflowStepMeta struct {
	message string // message prefix
	room    string // Matrix room
}

type postMessageMatrixWorkflowStep struct {
	workflowStep
	postMessageMatrixWorkflowStepMeta
}

func (s postMessageMatrixWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	p := payload.(postMessageMatrixWorkflowPayload)
	msg := p.message

	// Append message specified in definition of this step as a prefix to the payload
	if s.message != "" {
		if p.message != "" {
			msg = fmt.Sprintf("%s %s", s.message, p.message)
		} else {
			msg = s.message
		}
	}
	_, err := e.client.SendText(id.RoomID(p.room), msg)
	if err != nil {
		e.log(err.Error())
	}

	return payload, err
}
