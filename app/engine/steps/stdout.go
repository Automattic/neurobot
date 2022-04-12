package steps

import (
	"fmt"
	"io"
	"os"

	botApp "neurobot/app/bot"
)

var out io.Writer = os.Stdout

type stdoutWorkflowStepRunner struct{}

func (runner stdoutWorkflowStepRunner) Run(p map[string]string) (map[string]string, error) {
	msg := p["Message"]
	if msg == "" {
		msg = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+msg)

	return p, nil
}

func NewStdOutRunner(meta map[string]string, botRegistry botApp.Registry) *stdoutWorkflowStepRunner {
	return &stdoutWorkflowStepRunner{}
}
