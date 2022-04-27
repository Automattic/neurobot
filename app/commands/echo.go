package commands

import (
	"neurobot/model/command"
	"strings"
)

type echo struct {
	payload map[string]string
}

func (e *echo) Valid() bool {
	if len(e.payload["message"]) == 0 {
		return false
	}

	return true
}

func (e *echo) UsageHints() string {
	return "Usage: `!echo some text`"
}

func (e *echo) WorkflowPayload() map[string]string {
	return e.payload
}

// NewEcho returns an instance of ECHO command that handles its validation and set defaults for payload
func NewEcho(comm *command.Command) Command {
	payload := make(map[string]string)

	// convert args map into an args slice
	args := make([]string, 0, len(comm.Args))
	for _, v := range comm.Args {
		args = append(args, v)
	}

	payload["message"] = strings.TrimSpace(strings.Join(args, " "))
	payload["room"] = comm.Meta["room"]

	return &echo{
		payload: payload,
	}
}
