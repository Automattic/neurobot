package commands

import (
	"neurobot/model/command"
	"neurobot/model/payload"
	"strings"
)

type echo struct {
	payload payload.Payload
}

func (e *echo) Valid() bool {
	if len(e.payload.Message) == 0 {
		return false
	}

	return true
}

func (e *echo) UsageHints() string {
	return "Usage: `!echo some text`"
}

func (e *echo) WorkflowPayload() payload.Payload {
	return e.payload
}

// NewEcho returns an instance of ECHO command that handles its validation and set defaults for payload
func NewEcho(comm *command.Command) Command {
	var payload payload.Payload

	// convert args map into an args slice
	args := make([]string, 0, len(comm.Args))
	for _, v := range comm.Args {
		args = append(args, v)
	}

	payload.Message = strings.TrimSpace(strings.Join(args, " "))
	payload.Room = comm.Meta["room"]

	return &echo{
		payload: payload,
	}
}
