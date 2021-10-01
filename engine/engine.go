package engine

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type Engine struct {
	debug               bool
	portWebhookListener string

	matrixhomeserver string
	matrixusername   string
	matrixpassword   string

	db  *sql.DB
	dbA dbAssist

	workflows map[uint64]workflow
	triggers  map[string]map[string]Trigger
	// steps     map[string]map[string]WorkflowStep

	client *mautrix.Client
}

type RunParams struct {
	Debug               bool
	PortWebhookListener string
	MatrixHomeServer    string
	MatrixUsername      string
	MatrixPassword      string
}

func (e *Engine) Startup() {
	// Initialize maps
	e.workflows = make(map[uint64]workflow)
	e.triggers = make(map[string]map[string]Trigger)
	e.triggers["webhook"] = make(map[string]Trigger)
	e.triggers["poller"] = make(map[string]Trigger)

	// Establish database connection
	err := e.loadDB()
	if err != nil {
		log.Fatal(err)
	}

	// Load registered workflows from the database and initialize the right triggers for them
	e.loadWorkflows()

	// Start Matrix client
	go func() {
		err = e.initMatrixClient()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// allow the matrix client to sync and be ready,
	// before we invoke run() on Engine
	time.Sleep(time.Second * 5)
}

func (e *Engine) ShutDown() {
	// Close database connection
	e.db.Close()
}

func (e *Engine) Run() {
	// Matrix client needs to be initialized first as any webhooks and/or scheduler
	// can try to run Matrix related tasks

	go e.runPoller()

	e.runWebhookListener()
}

func (e *Engine) log(m string) {
	if e.debug {
		fmt.Println(m)
	}
}

func (e *Engine) loadDB() (err error) {
	e.db, err = sql.Open("sqlite3", "./wfb.db")
	if err != nil {
		log.Println(err)
		return
	}

	e.dbA = *NewDBAssist(e.db, e.debug)

	// Handle schema
	return e.dbA.manageDBSchema()
}

func (e *Engine) registerWebhookTrigger(name string, description string, urlSuffix string) *webhookt {
	t := NewWebhookTrigger(name, description, urlSuffix, e)
	e.triggers["webhook"][urlSuffix] = t

	return t
}
func (e *Engine) registerRSSPollTrigger(name string, description string, url string, pollingInterval time.Duration) *pollt {
	t := NewRSSPollTrigger(name, description, url, pollingInterval, e)
	e.triggers["poller"][name] = t

	return t
}
func (e *Engine) registerHTTPPollTrigger(name string, description string, url string, pollingInterval time.Duration) *pollt {
	t := NewHTTPPollTrigger(name, description, url, pollingInterval, e)
	e.triggers["poller"][name] = t

	return t
}

func (e *Engine) loadWorkflows() {
	// @TODO: Read workflows from database and creates instances for triggers
	// Currently hardcoded

	// sample workflows
	workflows := []workflow{
		{
			id:          1,
			name:        "CURL Webhook",
			description: "Simple Curl based webhook listener",
		},
		{
			id:          2,
			name:        "WP.org News Blog RSS Feed Emailer",
			description: "Email workflow for RSS Feed of WP.org news blog",
		},
	}

	// register steps
	var x1, x2, x3 WorkflowStep
	x1 = sendEmailWorkflowStep{
		workflowStep: workflowStep{
			variety:     "sendEmail",
			name:        "Email HR",
			description: "Email HR about a new hire",
		},
		sendEmailWorkflowStepMeta: sendEmailWorkflowStepMeta{
			emailAddr: "hr@example.org",
		},
	}
	x2 = sendEmailWorkflowStep{
		workflowStep: workflowStep{
			variety:     "sendEmail",
			name:        "Email Subscribers of WP.org news blog",
			description: "Email folks about a new blog post on WP.org news blog",
		},
		sendEmailWorkflowStepMeta: sendEmailWorkflowStepMeta{
			emailAddr: "folks1@example.org,folks2@example.org,folks3@example.org",
		},
	}
	x3 = postMessageMatrixWorkflowStep{
		workflowStep: workflowStep{
			variety:     "postMatrixMessage",
			name:        "Inform Neso",
			description: "Let the team know about this event by posting to team's matrix room",
		},
		postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
			message: "Alert!",
			room:    "!tnmILBRzpgkBkwSyDY:matrix.test",
		},
	}

	// attach registered steps to sample workflows
	workflows[0].addWorkflowStep(x1)
	workflows[1].addWorkflowStep(x2)
	workflows[0].addWorkflowStep(x3)

	// register triggers and attach them to workflow
	e.registerWebhookTrigger("matticspace webhook", "", "mcsp").attachWorkflow(0)
	e.registerRSSPollTrigger("wp.org news rss", "", "https://wordpress.org/news/feed/", time.Hour).attachWorkflow(1)

	// load workflows in engine
	e.workflows[0] = workflows[0]
	e.workflows[1] = workflows[1]
}

func (e *Engine) runWebhookListener() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		e.log(fmt.Sprintf("Request received on webhook listener! %s", r.URL.Path))

		if !strings.HasPrefix(r.URL.Path, "/webhooks-listener/") {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		suffix := strings.TrimSuffix(
			strings.TrimPrefix(
				r.URL.Path,
				"/webhooks-listener/",
			),
			"/",
		)

		t, exists := e.triggers["webhook"][suffix]
		e.log(fmt.Sprintf("suffix: %s registered: %t", suffix, exists))
		if exists {
			// @TODO:
			// figure out what data do we have here
			// check in this order:
			// GET request
			// 		message param unless specified param=m, in which case read 'm' param
			// POST request
			//		data param

			keys, ok := r.URL.Query()["message"]
			if !ok || len(keys[0]) < 1 {
				http.Error(w, "400 bad request.", http.StatusBadRequest)
				return
			}

			t.process(keys[0])
			// t.process()
		} else {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
	})

	e.log(fmt.Sprintf("Starting webhook listener at port %s\n", e.portWebhookListener))
	if err := http.ListenAndServe(":"+e.portWebhookListener, nil); err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) runPoller() {
	e.log("Running pollers")
	for _, t := range e.triggers["poller"] {
		go func(t Trigger) {
			t.setup()
		}(t)
	}
}

func (e *Engine) initMatrixClient() (err error) {
	if e.debug {
		fmt.Println("Logging into", e.matrixhomeserver, "as", e.matrixusername)
	}

	e.client, err = mautrix.NewClient(e.matrixhomeserver, "", "")
	if err != nil {
		return
	}
	_, err = e.client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: e.matrixusername},
		Password:         e.matrixpassword,
		StoreCredentials: true,
	})
	if err != nil {
		return
	}

	fmt.Println("Matrix: Login successful!")

	syncer := e.client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)
	})

	err = e.client.Sync()
	if err != nil {
		return
	}

	return
}

func NewEngine(p RunParams) *Engine {
	e := Engine{}

	// setting run parameters
	e.debug = p.Debug
	e.portWebhookListener = p.PortWebhookListener
	e.matrixhomeserver = p.MatrixHomeServer
	e.matrixusername = p.MatrixUsername
	e.matrixpassword = p.MatrixPassword

	return &e
}
