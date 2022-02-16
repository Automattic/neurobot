package engine

type WorkflowStep interface {
	run(interface{}, *engine) (interface{}, error) // accepts payload from workflow and returns after modification (if desired)
}

type workflowStep struct {
	id          uint64
	name        string
	description string
	variety     string
	workflowID  uint64
}
