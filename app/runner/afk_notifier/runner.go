package afk_notifier

import (
	r "neurobot/app/runner"
	"neurobot/infrastructure/matrix"
	"neurobot/model/workflow"
)

type runner struct {
	matrixClient matrix.Client
}

func NewRunner(matrixClient matrix.Client) r.Runner {
	return &runner{matrixClient: matrixClient}
}

func (r *runner) Run(workflow workflow.Workflow, payload map[string]string) error {
	//TODO implement me
	panic("implement me")
}
