package engine

type Trigger interface {
	setup()
	process(interface{})
	finish(interface{})
}

type trigger struct {
	id          uint64
	variety     string
	name        string
	description string
	engine      *engine
	workflowID  uint64
}

func (t *trigger) setup() {}

func (t *trigger) finish(payload interface{}) {
	w := t.engine.workflows[t.workflowID]
	w.run(payload, t.engine)
}
