package afk_notifier

import (
	r "neurobot/app/runner"
	"neurobot/infrastructure/matrix"
	m "neurobot/model/message"
	"neurobot/model/room"
	"neurobot/model/workflow"
)

type runner struct {
	matrixClient matrix.Client
}

func NewRunner(matrixClient matrix.Client) r.Runner {
	return &runner{matrixClient: matrixClient}
}

func (r *runner) Run(workflow workflow.Workflow, payload map[string]string) error {
	roomID, err := room.NewID(payload["room"])
	if err != nil {
		return err
	}

	message := m.NewMarkdownMessage(payload["message"])

	return r.matrixClient.SendMessage(roomID, message)
}
