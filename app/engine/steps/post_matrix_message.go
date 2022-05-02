package steps

import (
	"errors"
	"fmt"
	botApp "neurobot/app/bot"
	"neurobot/infrastructure/matrix"
	"neurobot/model/message"
	"neurobot/model/payload"
	r "neurobot/model/room"

	"github.com/apex/log"
)

type postMatrixMessageWorkflowStepMeta struct {
	messagePrefix string // message prefix
	matrixRoom    string // Matrix room
	asBot         string // bot identifier, for matrix session
}

type postMatrixMessageWorkflowStepRunner struct {
	eid string
	postMatrixMessageWorkflowStepMeta
	botRegistry botApp.Registry
}

func (runner *postMatrixMessageWorkflowStepRunner) getMatrixClient() (mc matrix.Client, err error) {
	if runner.asBot == "" {
		// If no bot was specified, use the primary one.
		return runner.botRegistry.GetPrimaryClient()
	}

	return runner.botRegistry.GetClient(runner.asBot)
}

func (runner *postMatrixMessageWorkflowStepRunner) Run(p *payload.Payload) error {
	log.Log.WithFields(log.Fields{
		"executionID":  runner.eid,
		"workflowStep": "postMatrixMessage",
	}).Info("running workflow step")

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
	room := runner.matrixRoom
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

	return mc.SendMessage(roomID, message.NewMarkdownMessage(msg))
}

func NewPostMatrixMessageRunner(eid string, meta map[string]string, botRegistry botApp.Registry) *postMatrixMessageWorkflowStepRunner {
	var stepMeta postMatrixMessageWorkflowStepMeta
	var ok bool

	stepMeta.asBot, ok = meta["asBot"]
	if !ok {
		stepMeta.asBot = ""
	}

	stepMeta.matrixRoom, ok = meta["matrixRoom"]
	if !ok {
		stepMeta.matrixRoom = ""
	}

	stepMeta.messagePrefix, ok = meta["messagePrefix"]
	if !ok {
		stepMeta.messagePrefix = ""
	}

	return &postMatrixMessageWorkflowStepRunner{
		eid:                               eid,
		postMatrixMessageWorkflowStepMeta: stepMeta,
		botRegistry:                       botRegistry,
	}
}
