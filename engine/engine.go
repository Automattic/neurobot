package engine

import (
	"fmt"
	"log"
	"neurobot/infrastructure/database"
	"neurobot/infrastructure/event"
	"neurobot/model/bot"
	"neurobot/model/trigger"
	wf "neurobot/model/workflow"
	"strings"
	"sync"

	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	mautrixEvent "maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Engine interface {
	StartUp(MatrixClient, mautrix.Syncer)
	Run(wf.Workflow, map[string]string) error
	ShutDown()
	log(string)
}

type MatrixClient interface {
	Login(*mautrix.ReqLogin) (*mautrix.RespLogin, error)
	Sync() error
	ResolveAlias(alias id.RoomAlias) (resp *mautrix.RespAliasResolve, err error)
	SendText(roomID id.RoomID, text string) (*mautrix.RespSendEvent, error)
	SendMessageEvent(roomID id.RoomID, eventType mautrixEvent.Type, contentJSON interface{}, extra ...mautrix.ReqSendEvent) (resp *mautrix.RespSendEvent, err error)
	JoinRoom(roomIDorAlias string, serverName string, content interface{}) (resp *mautrix.RespJoinRoom, err error)
}

type engine struct {
	debug                bool
	workflowsDefTOMLFile string

	isMatrix         bool // Do we mean to run a matrix client?
	matrixServerName string
	matrixServerURL  string
	matrixusername   string
	matrixpassword   string

	db            db.Session
	eventBus      event.Bus
	botRepository bot.Repository

	workflows map[uint64]*workflow
	triggers  map[string]map[string]trigger.Trigger

	bots map[uint64]MatrixClient // All matrix client instances of bots

	client MatrixClient
}

type RunParams struct {
	EventBus             event.Bus
	BotRepository        bot.Repository
	Debug                bool
	Database             string
	WorkflowsDefTOMLFile string
	IsMatrix             bool
	MatrixServerName     string // domain in use, part of identity
	MatrixServerURL      string // actual URL to connect to, for a particular server
	MatrixUsername       string
	MatrixPassword       string
}

func (e *engine) StartUpLite() {
	e.log("Starting up engine..")

	// Initialize maps
	e.bots = make(map[uint64]MatrixClient)
	e.workflows = make(map[uint64]*workflow)
	e.triggers = make(map[string]map[string]trigger.Trigger)
	e.triggers["webhook"] = make(map[string]trigger.Trigger)

	// Establish database connection
	e.log("Attempting to establish database connection..")
	err := e.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	// Check for workflows defined in TOML
	e.handleTOMLDefinitions()

	// Load registered workflows from the database and initialize the right triggers for them
	e.log("Loading data...")
	e.loadData()

	go e.eventBus.Subscribe(event.TriggerTopic(), func(event interface{}) {
		var t trigger.Trigger

		switch event.(type) {
		default:
			return
		case trigger.Trigger:
			t = event.(trigger.Trigger)
		}

		workflow := e.workflows[t.WorkflowID]
		workflow.run(t.Payload, e)
	})

	e.log("Finished starting up engine.")
}

func (e *engine) Run(w wf.Workflow, payload map[string]string) error {
	workflow := e.workflows[w.ID]
	workflow.run(payload, e)
	return nil
}

func (e *engine) StartUp(mc MatrixClient, s mautrix.Syncer) {
	e.StartUpLite()

	// Start Matrix client, if desired
	// Note: Matrix client needs to be initialized early as a trigger can try to run Matrix related tasks
	if e.isMatrix {
		e.log("Starting up Matrix client(s)..")

		var wg sync.WaitGroup
		wg.Add(2)

		// This creates the matrix instance of the main/god bot
		go func() {
			defer wg.Done()

			err := e.initMatrixClient(mc, s)
			if err != nil {
				log.Fatal(err)
			}
			e.log("Finished loading up God bot.")
		}()

		// This creates the matrix instances of all other bots
		go func() {
			defer wg.Done()

			err := e.wakeUpMatrixBots()
			if err != nil {
				log.Fatal(err) // fatal error for now
			}
			e.log("Finished waking up all Matrix bots.")
		}()

		// allow the matrix client(s) to sync and be ready,
		wg.Wait()
		e.log("Engine's matrix start up finished.")
	}
}

func (e *engine) ShutDown() {
	// Close database connection
	e.db.Close()
}

func (e *engine) log(m string) {
	if e.debug {
		fmt.Println(m)
	}
}

func (e *engine) loadDB() (err error) {
	// Use upper.io ORM now
	e.db, err = database.MakeDatabaseSession()
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	err = database.Migrate(e.db)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	// Set database logging to Errors only when debug:false
	if !e.debug {
		db.LC().SetLevel(db.LogLevelError)
	}

	return
}

func (e *engine) loadData() {
	// load workflows
	workflows, err := getConfiguredWorkflows(e.db)
	if err != nil {
		log.Fatalf("Error loading workflows from database: %s", err)
	}
	for _, w := range workflows {
		// copy over the value in a separate variable because we need to store a pointer
		// w gets assigned a different value with every iteration, which modifies all values if address of w is taken directly
		instance := w
		e.workflows[w.id] = &instance
	}

	// load workflow steps
	steps, err := getConfiguredWFSteps(e.db)
	if err != nil {
		log.Fatalf("Error loading workflow steps from database: %s", err)
	}
	for _, ws := range steps {
		switch ws := ws.(type) {
		case *postMessageMatrixWorkflowStep:
			fmt.Printf("Adding %s to workflow #%d\n", ws.name, ws.workflowID)
			e.workflows[ws.workflowID].addWorkflowStep(ws)
		case *stdoutWorkflowStep:
			fmt.Printf("Adding %s to workflow #%d\n", ws.name, ws.workflowID)
			e.workflows[ws.workflowID].addWorkflowStep(ws)
		case *sendEmailWorkflowStep:
			fmt.Printf("Adding %s to workflow #%d\n", ws.name, ws.workflowID)
			e.workflows[ws.workflowID].addWorkflowStep(ws)
		}
	}
}

func (e *engine) handleTOMLDefinitions() {
	if err := parseTOMLDefs(e); err != nil {
		log.Fatal(err)
	}
}

func (e *engine) initMatrixClient(c MatrixClient, s mautrix.Syncer) (err error) {
	e.client = c

	e.log(fmt.Sprintf("Matrix: Logging into %s as %s", e.matrixServerName, e.matrixusername))

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

	e.log("Matrix: Login successful!")

	syncer := s.(*mautrix.DefaultSyncer)
	syncer.OnEventType(mautrixEvent.EventMessage, func(source mautrix.EventSource, evt *mautrixEvent.Event) {
		e.log(fmt.Sprintf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body))
	})
	syncer.OnEventType(mautrixEvent.StateMember, func(source mautrix.EventSource, evt *mautrixEvent.Event) {
		if membership, ok := evt.Content.Raw["membership"]; ok {
			if membership == "invite" {
				e.log(fmt.Sprintf("neurobot got invitation for %s\n", evt.RoomID))

				// ensure the invitation is for a room within our homeserver only
				matrixHSHost := strings.Split(e.matrixServerName, ":")[0] // remove protocol and port info to get just the hostname
				if strings.Split(evt.RoomID.String(), ":")[1] == matrixHSHost {
					// join the room
					_, err := e.client.JoinRoom(evt.RoomID.String(), "", "")
					if err != nil {
						e.log(fmt.Sprintf("neurobot couldn't join the invitation: %s", evt.RoomID))
					} else {
						e.log("neurobot accepted invitation, if it wasn't accepted already")
					}
				} else {
					e.log(fmt.Sprintf("neurobot whaat? %v", strings.Split(evt.RoomID.String(), ":")))
				}
			}
		}
	})

	// Fire 'sync' in another go routine since its blocking
	go func() {
		e.log(e.client.Sync().Error())
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
		return fmt.Errorf("one or more bots could not wake up. ids: %v", failedWakeUps)
	}

	return nil
}

func NewEngine(p RunParams) *engine {
	e := engine{}

	// setting run parameters
	e.debug = p.Debug
	e.workflowsDefTOMLFile = p.WorkflowsDefTOMLFile
	e.isMatrix = p.IsMatrix
	e.matrixServerName = p.MatrixServerName
	e.matrixServerURL = p.MatrixServerURL
	e.matrixusername = p.MatrixUsername
	e.matrixpassword = p.MatrixPassword
	e.eventBus = p.EventBus
	e.botRepository = p.BotRepository

	return &e
}
