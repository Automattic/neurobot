package engine

type mockWorkflowStep struct {
	impact string
}

func (m *mockWorkflowStep) run(payload string, e *Engine) string {
	return payload + m.impact
}

func NewMockWorkflowStep(impact string) *mockWorkflowStep {
	return &mockWorkflowStep{impact: impact}
}
