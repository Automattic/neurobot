package engine

type Trigger struct {
	id          uint64
	variety     string
	name        string
	description string
	workflowID  uint64
	payload     payloadData
	meta        map[string]string
}

func (t *Trigger) GetWorkflowId() uint64 {
	return t.workflowID
}

func (t *Trigger) GetPayload() payloadData {
	return t.payload
}

func (t *Trigger) SetPayload(payload payloadData) {
	t.payload = payload
}
