package steps

import (
	"fmt"
	"io"
	"os"

	"neurobot/model/payload"

	"github.com/apex/log"
)

var out io.Writer = os.Stdout

type stdoutWorkflowStepRunner struct {
	eid string
}

func (runner *stdoutWorkflowStepRunner) Run(p *payload.Payload) error {
	log.Log.WithFields(log.Fields{
		"executionID":  runner.eid,
		"workflowStep": "stdout",
	}).Info("running workflow step")

	msg := p.Message
	if msg == "" {
		msg = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+msg)

	return nil
}

func NewStdOutRunner(eid string, meta map[string]string) *stdoutWorkflowStepRunner {
	return &stdoutWorkflowStepRunner{
		eid: eid,
	}
}
