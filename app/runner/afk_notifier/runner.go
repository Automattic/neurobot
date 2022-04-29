package afk_notifier

import (
	r "neurobot/app/runner"
	"neurobot/infrastructure/matrix"
	m "neurobot/model/message"
	"neurobot/model/payload"
	"neurobot/model/room"
	"neurobot/model/workflow"
)

type runner struct {
	matrixClient matrix.Client
}

func NewRunner(matrixClient matrix.Client) r.Runner {
	return &runner{matrixClient: matrixClient}
}

func (r *runner) Run(eid string, workflow workflow.Workflow, payload payload.Payload) error {
	roomID, err := room.NewID(payload.Room)
	if err != nil {
		return err
	}

	message := m.NewMarkdownMessage(payload.Message)

	return r.matrixClient.SendMessage(roomID, message)
}
