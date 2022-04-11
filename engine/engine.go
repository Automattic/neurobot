package engine

import (
	"fmt"
	b "neurobot/app/bot"
	"neurobot/model/bot"
	wf "neurobot/model/workflow"
	wfs "neurobot/model/workflowstep"

	"github.com/apex/log"

	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	mautrixEvent "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Engine interface {
	StartUp(MatrixClient, mautrix.Syncer)
	Run(wf.Workflow, map[string]string) error
}

type MatrixClient interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	Sync() error
	ResolveAlias(alias id.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
	SendText(roomID id.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID id.RoomID, eventType mautrixEvent.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	JoinRoom(roomIDorAlias string, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
}

type WorkflowStepRunner interface {
	run(map[string]string) (map[string]string, error) // accepts payload and returns after modification (if desired)
}

type engine struct {
	debug bool

	matrixServerName string
	matrixServerURL  string
	matrixusername   string
	matrixpassword   string

	db            db.Session
	botRepository bot.Repository
	botRegistry   b.Registry

	workflowRepository     wf.Repository
	workflowStepRepository wfs.Repository

	workflows map[uint64]*wf.Workflow
	bots      map[uint64]MatrixClient // All matrix client instances of bots

	client MatrixClient
}

type RunParams struct {
	BotRepository          bot.Repository
	BotRegistry            b.Registry
	WorkflowRepository     wf.Repository
	WorkflowStepRepository wfs.Repository

	Debug            bool
	MatrixServerName string // domain in use, part of identity
	MatrixServerURL  string // actual URL to connect to, for a particular server
	MatrixUsername   string
	MatrixPassword   string
}

func (e *engine) Run(w wf.Workflow, payload map[string]string) error {
	logger := log.Log

	// loop through all the steps inside of the workflow
	steps, err := e.workflowStepRepository.FindByWorkflowID(w.ID)
	if err != nil {
		return fmt.Errorf("error fetching workflow steps while running workflow %d : %w", w.ID, err)
	}

	var runners []WorkflowStepRunner

	for _, s := range steps {
		switch s.Variety {
		case "postMatrixMessage":
			runners = append(runners, NewPostMatrixMessageRunner(s.Meta, e.botRegistry))
		case "stdOut":
			runners = append(runners, NewStdOutRunner(s.Meta, e.botRegistry))
		}
	}

	for _, r := range runners {
		payload, err = r.run(payload)
		if err != nil {
			// For now, we don't halt the workflow if a workflow step encounters an error
			logger.WithError(err).WithFields(log.Fields{
				"Identifier": w.Identifier,
			}).Info("workflow step execution error")
		}
	}

	return nil
}

func (e *engine) StartUp(mc MatrixClient, s mautrix.Syncer) {
	logger := log.Log
	logger.Info("Starting up engine")

	// Load registered workflows from the database and initialize the right triggers for them
	logger.Info("Loading data")
	e.loadData()

	logger.Info("Finished starting up engine.")
}

func (e *engine) loadData() {
	logger := log.Log

	// load workflows
	workflows, err := e.workflowRepository.FindActive()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load workflows from database")
	}
	for _, w := range workflows {
		// copy over the value in a separate variable because we need to store a pointer
		// w gets assigned a different value with every iteration, which modifies all values if address of w is taken directly
		instance := w
		e.workflows[w.ID] = &instance
	}
}

func NewEngine(p RunParams) *engine {
	e := engine{}

	// setting run parameters
	e.debug = p.Debug
	e.matrixServerName = p.MatrixServerName
	e.matrixServerURL = p.MatrixServerURL
	e.matrixusername = p.MatrixUsername
	e.matrixpassword = p.MatrixPassword
	e.botRepository = p.BotRepository
	e.botRegistry = p.BotRegistry
	e.workflowRepository = p.WorkflowRepository
	e.workflowStepRepository = p.WorkflowStepRepository

	// initialize maps
	e.bots = make(map[uint64]MatrixClient)
	e.workflows = make(map[uint64]*wf.Workflow)

	return &e
}
