package engine

import (
	"testing"

	"maunium.net/go/mautrix"
)

func TestGenericGetMatrixClient(t *testing.T) {
	tables := []struct {
		homeserver string
		isErr      bool
	}{
		{
			homeserver: "https://example.org",
			isErr:      false,
		},
		{
			homeserver: "example.org",
			isErr:      false,
		},
		{
			homeserver: " http://foo.com",
			isErr:      true,
		},
	}
	for _, table := range tables {
		c, err := getMatrixClient(table.homeserver)
		if err != nil {
			if !table.isErr {
				t.Error("error thrown when it shouldn't have")
			}

		} else {
			if table.isErr {
				t.Error("error not thrown when it should have")
			}

			if _, ok := c.(*mautrix.Client); !ok {
				t.Error("mautrix client wasn't returned")
			}
		}
	}
}

func TestGetMatrixClient(t *testing.T) {
	// override getMatrixClient() for returning a mock instance of matrix client even if "asBot" is defined
	// and we will simply check who does the mock instance belong to
	var org = getMatrixClient
	defer func() {
		getMatrixClient = org
	}()
	getMatrixClient = func(homeserver string) (MatrixClient, error) {
		return NewMockMatrixClient("bot"), nil
	}

	tables := []struct {
		asBot        string
		clientOrigin string
	}{
		// When bot identifier isn't specified, use the matrix client provided by engine
		{
			asBot:        "",
			clientOrigin: "engine",
		},
		// When bot identifier is specified, use the matrix client instatiated by the particular bot credentials
		{
			asBot:        "bot_something",
			clientOrigin: "bot1",
		},
		// When bot identifier is specified, use the matrix client instatiated by the particular bot credentials
		{
			asBot:        "bot_afk",
			clientOrigin: "bot2",
		},
	}

	for _, table := range tables {
		// setup db row in bots table
		dbs, dbs2 := setUp()
		defer tearDown(dbs, dbs2)

		// setup mock engine
		e := NewMockEngine()
		e.db = dbs
		e.bots = make(map[uint64]MatrixClient)
		e.bots[1] = NewMockMatrixClient("bot1")
		e.bots[2] = NewMockMatrixClient("bot2")

		// get step instance
		s := &postMessageMatrixWorkflowStep{
			workflowStep: workflowStep{
				id:      1,
				name:    "Post message to Matrix",
				variety: "postMatrixMessage",
			},
			postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
				messagePrefix: "",
				room:          "",
				asBot:         table.asBot,
			},
		}

		// finally, call the function we are testing
		client, err := s.getMatrixClient(e)
		if err != nil {
			t.Error("couldn't get a matrix client")
		}

		got := client.(*mockMatrixClient).instantiatedBy
		if table.clientOrigin != got {
			t.Errorf("right matrix client instance wasn't used. expected:%s got:%s", table.clientOrigin, got)
		}
	}
}

func TestPostMessageMatrixWorkflowStep(t *testing.T) {
	// override getMatrixClient() for returning a mock instance of matrix client even if "asBot" is defined
	// and we will simply check who does the mock instance belong to
	var org = getMatrixClient
	defer func() {
		getMatrixClient = org
	}()

	botMatrixClient, _ := getMockMatrixClient("doesntmatter") // homeserver arg is only used for testing code path that returns error when homeserver is invalid
	getMatrixClient = func(hs string) (MatrixClient, error) {
		return botMatrixClient, nil
	}

	tables := []struct {
		stepPrefixMessage string
		payload           string
		messageSent       string
		isError           bool
		asBot             string
		homeserver        string
	}{
		{
			stepPrefixMessage: "Test!",
			payload:           "Message!",
			messageSent:       "Test!\nMessage!",
			isError:           false,
			asBot:             "",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "Test!",
			payload:           "Message!",
			messageSent:       "Test!\nMessage!",
			isError:           false,
			asBot:             "bot_something",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "Message!",
			messageSent:       "Message!",
			isError:           false,
			asBot:             "",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "Message!",
			messageSent:       "Message!",
			isError:           false,
			asBot:             "bot_something",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "Hello: ",
			payload:           "",
			messageSent:       "Hello: ",
			isError:           false,
			asBot:             "",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "Hello: ",
			payload:           "",
			messageSent:       "Hello: ",
			isError:           false,
			asBot:             "bot_something",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "",
			messageSent:       "",
			isError:           false,
			asBot:             "",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "",
			messageSent:       "",
			isError:           false,
			asBot:             "bot_something",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "throwerr",
			messageSent:       "",
			isError:           true,
			asBot:             "",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "",
			payload:           "throwerr",
			messageSent:       "",
			isError:           true,
			asBot:             "bot_something",
			homeserver:        "https://example.com",
		},
		{
			stepPrefixMessage: "Test!",
			payload:           "Message!",
			messageSent:       "Test!\nMessage!",
			isError:           true,
			asBot:             "bot_something",        // otherwise e.client will be used
			homeserver:        " https://example.com", // invalid url to test returning of an error
		},
		{
			stepPrefixMessage: "Test!",
			payload:           "Message!",
			messageSent:       "Test!\nMessage!",
			isError:           true,
			asBot:             "bot_nonexistent",      // otherwise e.client will be used
			homeserver:        " https://example.com", // invalid url to test returning of an error
		},
	}

	for _, table := range tables {
		// setup db row in bots table
		dbs, dbs2 := setUp()

		e := NewMockEngine()
		e.db = dbs
		e.matrixhomeserver = table.homeserver
		e.bots = make(map[uint64]MatrixClient)
		e.bots[1] = botMatrixClient

		s := &postMessageMatrixWorkflowStep{
			workflowStep: workflowStep{
				id:      1,
				name:    "Post message to Matrix",
				variety: "postMatrixMessage",
			},
			postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
				messagePrefix: table.stepPrefixMessage,
				room:          "RRRR",
				asBot:         table.asBot,
			},
		}

		_, err := s.run(payloadData{Message: table.payload}, e)
		if err != nil {
			if !table.isError {
				t.Errorf("throwing error when it should not. payload: (%s) stepPrefixMessage: (%s)", table.payload, table.stepPrefixMessage)
			}
		}

		if !table.isError {
			if table.asBot != "" {
				if !botMatrixClient.(*mockMatrixClient).WasMessageSent(table.messageSent) {
					t.Errorf("Matrix message was not posted booooo")
				}
			} else {
				if !e.client.(*mockMatrixClient).WasMessageSent(table.messageSent) {
					t.Errorf("Matrix message was not posted. \n%v\n payload: (%s) stepPrefixMessage: (%s)", table, table.payload, table.stepPrefixMessage)
				}
			}
		}

		tearDown(dbs, dbs2)
	}

}
