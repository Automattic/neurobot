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

func (s postMessageMatrixWorkflowStep) run(payload string, e *Engine) string {
	// Append message specified in definition of this step as a prefix to the payload
	msg := fmt.Sprintf("%s\n%s", s.message, payload)
	_, err := e.client.SendText(id.RoomID(s.room), msg)
	if err != nil {
		fmt.Println(err)
	}

	return payload
}

func NewPostMessageMatrixWorkflowStep(name string, description string, payload string, message string, room string) *postMessageMatrixWorkflowStep {
	return &postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			variety:     "postMessageMatrix",
			name:        name,
			description: description,
			payload:     payload,
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			message: message,
			room:    room,
		},
	}
}
