package engine

import (
	"bytes"
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

	e := NewMockEngine()
	for _, table := range tables {
		s := &stdoutWorkflowStep{}
		s.run(map[string]string{"Message": table.input}, e)

		got := out.(*bytes.Buffer).String()
		if got != table.output+"\n" {
			t.Errorf("stdout message logged did not match. expected: (%s) got: (%s)", table.output, got)
		}

		out.(*bytes.Buffer).Reset()
	}
}
