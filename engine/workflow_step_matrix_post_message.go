package engine

import (
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

type postMessageMatrixWorkflowPayload struct {
	message string
	room    string
}

type postMessageMatrixWorkflowStepMeta struct {
	messagePrefix string // message prefix
	room          string // Matrix room
	asBot         string // bot identifier, for matrix session
}

type postMessageMatrixWorkflowStep struct {
	workflowStep
	postMessageMatrixWorkflowStepMeta
}

var getMatrixClient = func(homeserver string) (MatrixClient, error) {
	mc, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	return mc, nil
}

func (s postMessageMatrixWorkflowStep) getMatrixClient(e *engine) (mc MatrixClient, err error) {
	if s.asBot != "" {

		b, err := getBot(e.db, s.asBot)
		if err != nil {
			return nil, err
		}

		mc, err := getMatrixClient(e.matrixhomeserver)
		if err != nil {
			return nil, err
		}

		_, err = mc.Login(&mautrix.ReqLogin{
			Type:             "m.login.password",
			Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.Username},
			Password:         b.Password,
			StoreCredentials: true,
		})
		if err != nil {
			return nil, err
		}

		return mc, nil
	}

	return e.client, nil
}

func (s postMessageMatrixWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	p := payload.(postMessageMatrixWorkflowPayload)
	msg := p.message

	// Append message specified in definition of this step as a prefix to the payload
	if s.messagePrefix != "" {
		if p.message != "" {
			msg = fmt.Sprintf("%s\n%s", s.messagePrefix, p.message)
		} else {
			msg = s.messagePrefix
		}
	}

	mc, err := s.getMatrixClient(e)
	if err != nil {
		return nil, err
	}

	_, err = mc.SendText(id.RoomID(p.room), msg)
	if err != nil {
		return payload, err
	}

	return payload, nil
}
