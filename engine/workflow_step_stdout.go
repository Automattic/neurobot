package engine

import (
	"fmt"
	"io"
	"os"
)

var out io.Writer = os.Stdout

type stdoutWorkflowPayload struct {
	message string
}

type stdoutWorkflowStep struct {
	workflowStep
}

func (s stdoutWorkflowStep) run(p map[string]string, e *engine) (map[string]string, error) {
	if p["Message"] == "" {
		p["Message"] = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+p["Message"])

	return p, nil
}
