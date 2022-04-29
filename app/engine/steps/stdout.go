package steps

import (
	"fmt"
	"io"
	"os"

	botApp "neurobot/app/bot"
	"neurobot/model/payload"
)

var out io.Writer = os.Stdout

type stdoutWorkflowStepRunner struct{}

func (runner *stdoutWorkflowStepRunner) Run(p *payload.Payload) error {
	msg := p.Message
	if msg == "" {
		msg = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+msg)

	return nil
}

func NewStdOutRunner(meta map[string]string, botRegistry botApp.Registry) *stdoutWorkflowStepRunner {
	return &stdoutWorkflowStepRunner{}
}
