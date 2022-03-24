package trigger

type Trigger struct {
	ID          uint64
	Variety     string
	Name        string
	Description string
	WorkflowID  uint64
	Payload     map[string]string
	Meta        map[string]string
}

func (t *Trigger) GetWorkflowId() uint64 {
	return t.WorkflowID
}

func (t *Trigger) GetPayload() map[string]string {
	return t.Payload
}

func (t *Trigger) SetPayload(payload map[string]string) {
	t.Payload = payload
}
