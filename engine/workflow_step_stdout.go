package engine

import (
	"fmt"
	"io"
	"os"
)

var out io.Writer = os.Stdout

type stdoutWorkflowStep struct {
	workflowStep
}

func (s stdoutWorkflowStep) run(payload string, e *engine) (string, error) {
	if payload == "" {
		payload = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+payload)

	return payload, nil
}
