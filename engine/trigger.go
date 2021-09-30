package engine

type Trigger interface {
	attachWorkflow(uint64)
	setup()
	process(interface{})
	finish(string)
}

type trigger struct {
	variety     string
	name        string
	description string
	engine      *Engine
	workflows   []uint64 // a trigger can start multiple workflows
}

func (t *trigger) attachWorkflow(id uint64) {
	t.workflows = append(t.workflows, id)
}
func (t *trigger) finish(payload string) {
	// loop through all workflows meant for this trigger and run them
	for _, workflow := range t.workflows {
		w := t.engine.workflows[workflow]
		w.run(payload, t.engine)
	}
}
