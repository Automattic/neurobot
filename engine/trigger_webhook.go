package engine

type webhooktMeta struct {
	urlSuffix string
}

type webhookt struct {
	trigger
	webhooktMeta
}

func (t *webhookt) process(p payloadData) {
	t.finish(p)
}
