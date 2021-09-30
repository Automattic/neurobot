package engine

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Engine struct {
	debug               bool
	portWebhookListener string
	// portMatrixClient    string

	db  *sql.DB
	dbA dbAssist

	workflows map[uint64]workflow
	triggers  map[string]map[string]Trigger
	// steps     map[string]map[string]WorkflowStep

	// client mautrix.Client
}

type RunParams struct {
	Debug               bool
	PortWebhookListener string
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

	e.initMatrixClient()
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
	var x1, x2 WorkflowStep
	x1 = sendEmailWorkflowStep{
		workflowStep: workflowStep{
			variety:     "sendEmail",
			name:        "Email HR",
			description: "Email HR about a new hire",
			payload:     "",
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
			payload:     "",
		},
		sendEmailWorkflowStepMeta: sendEmailWorkflowStepMeta{
			emailAddr: "folks1@example.org,folks2@example.org,folks3@example.org",
		},
	}

	// attach registered steps to sample workflows
	workflows[0].addWorkflowStep(x1)
	workflows[1].addWorkflowStep(x2)

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

func (e *Engine) initMatrixClient() error {
	// matrixClients := clients.New(e.db, e.client)
	// if err := matrixClients.Start(); err != nil {
	// 	log.WithError(err).Panic("Failed to start up clients")
	// }

	// setup(e, http.DefaultServeMux, http.DefaultClient)
	// log.Fatal(http.ListenAndServe(e.BindAddress, nil))

	return nil // @TODO tmp fix
}

func NewEngine(defaults RunParams) *Engine {
	e := Engine{}

	// setting run parameters
	e.debug = defaults.Debug
	e.portWebhookListener = defaults.PortWebhookListener

	return &e
}
