package engine

type WorkflowStep interface {
	run(string, *Engine) string // accepts payload from workflow and returns after modification (if desired)
}

type workflowStep struct {
	variety     string
	name        string
	description string
	payload     string
}
