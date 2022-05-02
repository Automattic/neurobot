package commands

import (
	"fmt"
	"neurobot/model/command"
	"neurobot/model/payload"
)

const unrecognizedCommandUsageHint = "😱 Unrecognized command `!%s` please consult documentation"

type unrecognized struct {
	commandName string
	payload     payload.Payload
}

func (u *unrecognized) Valid() bool {
	return false
}

func (u *unrecognized) UsageHints() string {
	return fmt.Sprintf(unrecognizedCommandUsageHint, u.commandName)
}

func (u *unrecognized) WorkflowPayload() payload.Payload {
	return u.payload
}

// NewUnrecognized returns an instance of Command meant to be a catch-all for undefined commands invoked
func NewUnrecognized(comm *command.Command) Command {
	var payload payload.Payload
	payload.Room = comm.Meta["room"]

	return &unrecognized{
		commandName: comm.Name,
		payload:     payload,
	}
}
