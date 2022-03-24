package engine

type Trigger struct {
	id          uint64
	variety     string
	name        string
	description string
	workflowID  uint64
	payload     map[string]string
	meta        map[string]string
}

func (t *Trigger) GetWorkflowId() uint64 {
	return t.workflowID
}

func (t *Trigger) GetPayload() map[string]string {
	return t.payload
}

func (t *Trigger) SetPayload(payload map[string]string) {
	t.payload = payload
}
