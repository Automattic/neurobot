package engine

type WorkflowStep interface {
	run(string, *engine) (string, error) // accepts payload from workflow and returns after modification (if desired)
}

type workflowStep struct {
	id          uint64
	name        string
	description string
	variety     string
	workflow_id uint64
}
