package engine

type webhooktMeta struct {
	urlSuffix string
}

type webhookt struct {
	trigger
	webhooktMeta
}

func (t *webhookt) setup() {
	// no setup required for webhook triggers
}
func (t *webhookt) process(payload interface{}) {
	t.finish("Hello " + payload.(string))
}

func NewWebhookTrigger(name string, description string, urlSuffix string, engine *Engine) *webhookt {
	return &webhookt{
		trigger: trigger{
			variety:     "webhook",
			name:        name,
			description: description,
			engine:      engine,
		},
		webhooktMeta: webhooktMeta{
			urlSuffix: urlSuffix,
		},
	}
}
