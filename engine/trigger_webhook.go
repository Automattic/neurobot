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
