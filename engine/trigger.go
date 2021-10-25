package engine

type Trigger interface {
	setup()
	process(interface{})
	finish(string)
}

type trigger struct {
	id          uint64
	variety     string
	name        string
	description string
	engine      *engine
	workflows   []uint64 // a trigger can start multiple workflows
}

func (t *trigger) setup() {}

func (t *trigger) finish(payload string) {
	// loop through all workflows meant for this trigger and run them
	for _, workflowID := range t.workflows {
		// Get workflow instance and run() it
		w := t.engine.workflows[workflowID]
		w.run(payload, t.engine)
	}
}
