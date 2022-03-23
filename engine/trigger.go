package engine

type Trigger interface {
	process(payloadData)
	finish(payloadData)
	GetWorkflowId() uint64
	GetPayload() payloadData
	SetPayload(payload payloadData)
}

type trigger struct {
	id          uint64
	variety     string
	name        string
	description string
	engine      *engine
	workflowID  uint64
	payload     payloadData
}

func (t *trigger) GetWorkflowId() uint64 {
	return t.workflowID
}

func (t *trigger) GetPayload() payloadData {
	return t.payload
}

func (t *trigger) SetPayload(payload payloadData) {
	t.payload = payload
}

func (t *trigger) finish(p payloadData) {
	w := t.engine.workflows[t.workflowID]
	w.run(p, t.engine)
}
