package steps

import (
	"errors"
	"fmt"
	botApp "neurobot/app/bot"
	"neurobot/infrastructure/matrix"
	"neurobot/model/message"
	"neurobot/model/payload"
	r "neurobot/model/room"
)

type postMatrixMessageWorkflowStepMeta struct {
	messagePrefix string // message prefix
	room          string // Matrix room
	asBot         string // bot identifier, for matrix session
}

type postMatrixMessageWorkflowStepRunner struct {
	postMatrixMessageWorkflowStepMeta
	botRegistry botApp.Registry
}

func (runner postMatrixMessageWorkflowStepRunner) getMatrixClient() (mc matrix.Client, err error) {
	if runner.asBot == "" {
		// If no bot was specified, use the primary one.
		return runner.botRegistry.GetPrimaryClient()
	}

	return runner.botRegistry.GetClient(runner.asBot)
}

func (runner postMatrixMessageWorkflowStepRunner) Run(p *payload.Payload) error {
	msg := p.Message

	// Append message specified in definition of this step as a prefix to the payload
	if runner.messagePrefix != "" {
		if p.Message != "" {
			msg = fmt.Sprintf("%s %s", runner.messagePrefix, p.Message)
		} else {
			msg = runner.messagePrefix
		}
	}

	// Override room defined in meta, if provided in payload
	room := runner.room
	if p.Room != "" {
		room = p.Room
	}

	// ensure we have data to work with
	if room == "" {
		return errors.New("no room to post")
	}
	if msg == "" {
		return errors.New("no message to post")
	}

	mc, err := runner.getMatrixClient()
	if err != nil {
		return err
	}

	roomID, err := r.NewID(room)
	if err != nil {
		return err
	}

	err = mc.SendMessage(roomID, message.NewMarkdownMessage(msg))

	return err
}

func NewPostMatrixMessageRunner(meta map[string]string, botRegistry botApp.Registry) *postMatrixMessageWorkflowStepRunner {
	var stepMeta postMatrixMessageWorkflowStepMeta
	var ok bool

	stepMeta.asBot, ok = meta["asBot"]
	if !ok {
		stepMeta.asBot = ""
	}

	stepMeta.room, ok = meta["room"]
	if !ok {
		stepMeta.room = ""
	}

	stepMeta.messagePrefix, ok = meta["messagePrefix"]
	if !ok {
		stepMeta.messagePrefix = ""
	}

	return &postMatrixMessageWorkflowStepRunner{
		postMatrixMessageWorkflowStepMeta: stepMeta,
		botRegistry:                       botRegistry,
	}
}
