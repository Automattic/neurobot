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

func (s stdoutWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	p := payload.(stdoutWorkflowPayload)
	if p.message == "" {
		p.message = "[Empty line]"
	}
	fmt.Fprintln(out, ">>"+p.message)

	return payload, nil
}
