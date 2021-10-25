package engine

import "testing"

func TestPostMessageMatrixWorkflowStep(t *testing.T) {
	tables := []struct {
		stepPrefixMessage string
		payload           string
		messageSent       string
		isError           bool
	}{
		{
			stepPrefixMessage: "Test!",
			payload:           "Message!",
			messageSent:       "Test! Message!",
			isError:           false,
		},
		{
			stepPrefixMessage: "",
			payload:           "Message!",
			messageSent:       "Message!",
			isError:           false,
		},
		{
			stepPrefixMessage: "Hello: ",
			payload:           "",
			messageSent:       "Hello: ",
			isError:           false,
		},
		{
			stepPrefixMessage: "",
			payload:           "",
			messageSent:       "",
			isError:           false,
		},
		{
			stepPrefixMessage: "",
			payload:           "throwerr",
			messageSent:       "",
			isError:           true,
		},
	}

	for _, table := range tables {
		m := NewMockMatrixClient()
		e := &engine{client: m}

		s := &postMessageMatrixWorkflowStep{
			workflowStep: workflowStep{
				id:      1,
				name:    "Post message to Matrix",
				variety: "postMatrixMessage",
			},
			postMessageMatrixWorkflowStepMeta: postMessageMatrixWorkflowStepMeta{
				message: table.stepPrefixMessage,
				room:    "RRRR",
			},
		}

		_, err := s.run(table.payload, e)
		if err != nil {
			if !table.isError {
				t.Errorf("throwing error when it should not. payload: (%s) stepPrefixMessage: (%s)", table.payload, table.stepPrefixMessage)
			}
		}

		if !table.isError && !m.(*mockMatrixClient).WasMessageSent(table.messageSent) {
			t.Errorf("Matrix message was not posted. payload: (%s) stepPrefixMessage: (%s)", table.payload, table.stepPrefixMessage)
		}
	}

}
