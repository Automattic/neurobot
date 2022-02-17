package engine

type Trigger interface {
	setup()
	process(payloadData)
	finish(payloadData)
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

func (t *trigger) finish(p payloadData) {
	w := t.engine.workflows[t.workflowID]
	w.run(p, t.engine)
}
