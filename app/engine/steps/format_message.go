package steps

import (
	"fmt"
	"neurobot/model/payload"

	"github.com/apex/log"
)

type formatMessageWorkflowStepMeta struct {
	variety string
}

type formatMessageWorkflowStepRunner struct {
	eid string
	formatMessageWorkflowStepMeta
}

func (runner *formatMessageWorkflowStepRunner) Run(p *payload.Payload) error {
	log.Log.WithFields(log.Fields{
		"executionID":  runner.eid,
		"workflowStep": "formatMessage",
	}).Info("running workflow step")

	switch runner.variety {
	case "appendUsersList":
		p.Message = p.Message + fmt.Sprintf("%+q", p.Users)
	}
	return nil
}

func NewFormatMessageRunner(eid string, meta map[string]string) *formatMessageWorkflowStepRunner {
	var stepMeta formatMessageWorkflowStepMeta
	stepMeta.variety, _ = meta["variety"]
	return &formatMessageWorkflowStepRunner{
		eid:                           eid,
		formatMessageWorkflowStepMeta: stepMeta,
	}
}
