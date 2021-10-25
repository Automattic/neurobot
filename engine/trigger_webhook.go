package engine

type webhooktMeta struct {
	urlSuffix string
}

type webhookt struct {
	trigger
	webhooktMeta
}

func (t *webhookt) process(payload interface{}) {
	t.finish(payload.(string))
}
