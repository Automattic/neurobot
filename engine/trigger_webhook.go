package engine

type webhooktMeta struct {
	urlSuffix string
}

type webhookt struct {
	trigger
	webhooktMeta
}
