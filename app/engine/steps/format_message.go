package steps

import (
	"fmt"
	"neurobot/model/payload"
)

type formatMessageWorkflowStepMeta struct {
	variety string
}

type formatMessageWorkflowStepRunner struct {
	formatMessageWorkflowStepMeta
}

func (runner *formatMessageWorkflowStepRunner) Run(p *payload.Payload) error {
	switch runner.variety {
	case "appendUsersList":
		p.Message = p.Message + fmt.Sprintf("%+q", p.Users)
	}
	return nil
}

func NewFormatMessageRunner(meta map[string]string) *formatMessageWorkflowStepRunner {
	var stepMeta formatMessageWorkflowStepMeta
	stepMeta.variety, _ = meta["variety"]
	return &formatMessageWorkflowStepRunner{stepMeta}
}
