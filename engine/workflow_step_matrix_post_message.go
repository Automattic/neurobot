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

func (s postMessageMatrixWorkflowStep) run(payload string, e *engine) string {
	// Append message specified in definition of this step as a prefix to the payload
	msg := fmt.Sprintf("%s\n%s", s.message, payload)
	_, err := e.client.SendText(id.RoomID(s.room), msg)
	if err != nil {
		fmt.Println(err)
	}

	return payload
}
