package commands

import (
	"fmt"
	"neurobot/model/command"
)

const unrecognizedCommandUsageHint = "😱 Unrecognized command `!%s` please consult documentation"

type unrecognized struct {
	commandName string
	payload     map[string]string
}

func (u *unrecognized) Valid() bool {
	return false
}

func (u *unrecognized) UsageHints() string {
	return fmt.Sprintf(unrecognizedCommandUsageHint, u.commandName)
}

func (u *unrecognized) WorkflowPayload() map[string]string {
	return u.payload
}

// NewUnrecognized returns an instance of Command meant to be a catch-all for undefined commands invoked
func NewUnrecognized(comm *command.Command) Command {
	payload := make(map[string]string)
	payload["room"] = comm.Meta["room"]

	return &unrecognized{
		commandName: comm.Name,
		payload:     payload,
	}
}
