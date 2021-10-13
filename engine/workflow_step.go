package engine

type WorkflowStep interface {
	run(string, *Engine) string // accepts payload from workflow and returns after modification (if desired)
}

type workflowStep struct {
	id          uint64
	name        string
	description string
	variety     string
	workflow_id uint64
}
