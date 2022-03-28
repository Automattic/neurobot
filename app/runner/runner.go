package runner

import (
	"neurobot/model/workflow"
)

// Runner runs a Workflow with an incoming payload.
type Runner interface {
	Run(workflow workflow.Workflow, payload map[string]string) error
}
