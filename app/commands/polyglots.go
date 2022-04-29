package commands

import (
	"fmt"
	"neurobot/model/command"
	"neurobot/model/payload"
)

type polyglots struct {
	payload payload.Payload
}

func (p *polyglots) Valid() bool {
	return true
}

func (p *polyglots) UsageHints() string {
	return "Usage: `!polyglots spanish`"
}

func (p *polyglots) WorkflowPayload() payload.Payload {
	return p.payload
}

// NewPolyglots returns an instance of POLYGLOTS command that handles its validation and set defaults for payload
func NewPolyglots(comm *command.Command) Command {
	var pg polyglots
	pg.payload.Room = comm.Meta["room"]

	// convert args map into an args slice
	args := make([]string, 0, len(comm.Args))
	for _, v := range comm.Args {
		args = append(args, v)
	}

	pg.payload.Context = make(map[string]string)
	pg.payload.Context["postParam1Name"] = "lang"
	pg.payload.Context["postParam1Value"] = args[0]

	pg.payload.Message = fmt.Sprintf("Folks who know '%s' and are online: ", args[0])

	return &pg
}
