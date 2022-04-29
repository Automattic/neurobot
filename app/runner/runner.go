package runner

import (
	"neurobot/model/payload"
	"neurobot/model/workflow"
)

// Runner runs a Workflow with an incoming payload.
type Runner interface {
	Run(eid string, workflow workflow.Workflow, payload payload.Payload) error
}
