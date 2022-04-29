package steps

import (
	"bytes"
	"neurobot/model/payload"
	"testing"
)

func TestStdoutWorkflowStep(t *testing.T) {
	tables := []struct {
		input  string
		output string
	}{
		{
			input:  "hello",
			output: ">>hello",
		},
		{
			input:  "",
			output: ">>[Empty line]",
		},
	}

	// hijack writer for this test
	backup := out
	out = new(bytes.Buffer)
	defer func() { out = backup }()

	for _, table := range tables {
		s := &stdoutWorkflowStepRunner{}

		s.Run(payload.Payload{
			Message: table.input,
		})

		got := out.(*bytes.Buffer).String()
		if got != table.output+"\n" {
			t.Errorf("stdout message logged did not match. expected: (%s) got: (%s)", table.output, got)
		}

		out.(*bytes.Buffer).Reset()
	}
}
