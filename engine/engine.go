package engine

import (
	"errors"
	"fmt"
	b "neurobot/app/bot"
	"neurobot/model/bot"
	wf "neurobot/model/workflow"
	wfs "neurobot/model/workflowstep"
	"strings"
	"sync"
	"time"

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

	isMatrix         bool // Do we mean to run a matrix client?
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
	IsMatrix         bool
	MatrixServerName string // domain in use, part of identity
	MatrixServerURL  string // actual URL to connect to, for a particular server
	MatrixUsername   string
	MatrixPassword   string
}

func (e *engine) StartUpLite() {
	logger := log.Log
	logger.Info("Starting up engine")

	// Initialize maps
	e.bots = make(map[uint64]MatrixClient)
	e.workflows = make(map[uint64]*wf.Workflow)

	// Load registered workflows from the database and initialize the right triggers for them
	logger.Info("Loading data")
	e.loadData()

	logger.Info("Finished starting up engine.")
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
	e.StartUpLite()

	// Start Matrix client, if desired
	// Note: Matrix client needs to be initialized early as a trigger can try to run Matrix related tasks
	if e.isMatrix {
		logger.Info("Starting up Matrix client(s)")

		var wg sync.WaitGroup
		wg.Add(2)

		// This creates the matrix instance of the main/god bot
		go func() {
			defer wg.Done()

			err := e.initMatrixClient(mc, s)
			if err != nil {
				logger.WithError(err).Fatal("Failed to init matrix client")
			}
			logger.Info("Finished starting primary bot")
		}()

		// This creates the matrix instances of all other bots
		go func() {
			defer wg.Done()

			err := e.wakeUpMatrixBots()
			if err != nil {
				logger.WithError(err).Fatal("Failed to wake up bots")
			}
			logger.Info("Finished waking up all Matrix bots")
		}()

		// allow the matrix client(s) to sync and be ready,
		wg.Wait()
		logger.Info("Engine's matrix start up finished")
	}
}

func (e *engine) loadData() {
	logger := log.Log

	// load workflows
	workflows, err := getConfiguredWorkflows(e.workflowRepository)
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

func (e *engine) initMatrixClient(c MatrixClient, s mautrix.Syncer) (err error) {
	logger := log.Log
	e.client = c

	start := time.Now()
	_, err = e.client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: e.matrixusername},
		Password:         e.matrixpassword,
		DeviceID:         "NEUROBOT",
		StoreCredentials: true,
	})
	if err != nil {
		return
	}

	logger.WithDuration(time.Since(start)).WithFields(log.Fields{
		"serverName": e.matrixServerName,
		"username":   e.matrixusername,
	}).Info("Logged in to homeserver")

	syncer := s.(*mautrix.DefaultSyncer)
	syncer.OnEventType(mautrixEvent.EventMessage, func(source mautrix.EventSource, evt *mautrixEvent.Event) {
		logger.WithFields(log.Fields{
			"sender": evt.Sender,
			"type":   evt.Type.String(),
			"body":   evt.Content.AsMessage().Body,
		}).Info("Matrix event")
	})
	syncer.OnEventType(mautrixEvent.StateMember, func(source mautrix.EventSource, evt *mautrixEvent.Event) {
		logger := log.WithFields(log.Fields{
			"room": evt.RoomID,
		})

		if membership, ok := evt.Content.Raw["membership"]; ok {
			if membership == "invite" {
				logger.Info("Neurobot was invited to room")

				// ensure the invitation is for a room within our homeserver only
				matrixHSHost := strings.Split(e.matrixServerName, ":")[0] // remove protocol and port info to get just the hostname
				if strings.Split(evt.RoomID.String(), ":")[1] == matrixHSHost {
					// join the room
					_, err := e.client.JoinRoom(evt.RoomID.String(), "", "")
					if err != nil {
						logger.WithError(err).Error("Neurobot could not accept invitation")
					} else {
						logger.Info("Neurobot accepted invitation, if it wasn't accepted already")
					}
				} else {
					logger.Warn("Neurobot was invited to a room in another homeserver")
				}
			}
		}
	})

	// Fire 'sync' in another go routine since its blocking
	go func() {
		err := e.client.Sync().Error()
		logger.WithField("error", err).Error("Sync failed")
	}()

	return
}

func (e *engine) wakeUpMatrixBots() (err error) {
	// load all bots one by one and accept any invitations within our own homeserver
	modelBots, err := e.botRepository.FindActive()
	if err != nil {
		return
	}

	// Convert model/bot to engine/bot
	// TODO: Remove once engine/bot has been replaced in favour of model/bot
	var bots []Bot
	for _, modelBot := range modelBots {
		bots = append(bots, MakeBotFromModelBot(modelBot))
	}

	// use waitgroup to wait for all bots' instances to be ready
	var wg sync.WaitGroup

	// collect bot IDs who error'd out
	var failedWakeUps []uint64

	// using go routines here to instantiate in parallel - rate limiting might become a problem with too many bots though
	for _, b := range bots {
		wg.Add(1)

		go func(b Bot) {
			defer wg.Done()

			if err := b.WakeUp(e); err != nil {
				failedWakeUps = append(failedWakeUps, b.ID)
			}
		}(b)

	}

	// wait for all bot instances to wake up
	wg.Wait()

	if len(failedWakeUps) > 0 {
		err = errors.New("one or more bots could not wake up")
		log.WithFields(log.Fields{
			"failedBots": failedWakeUps,
		}).Error("Failed to wake up bots")
		return err
	}

	return nil
}

func NewEngine(p RunParams) *engine {
	e := engine{}

	// setting run parameters
	e.debug = p.Debug
	e.isMatrix = p.IsMatrix
	e.matrixServerName = p.MatrixServerName
	e.matrixServerURL = p.MatrixServerURL
	e.matrixusername = p.MatrixUsername
	e.matrixpassword = p.MatrixPassword
	e.botRepository = p.BotRepository
	e.botRegistry = p.BotRegistry
	e.workflowRepository = p.WorkflowRepository
	e.workflowStepRepository = p.WorkflowStepRepository

	return &e
}
