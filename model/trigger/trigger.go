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
