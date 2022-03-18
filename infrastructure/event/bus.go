package event

type Bus interface {
	Publish(topic Topic, event interface{})
	Subscribe(topic Topic, handler func(event interface{}))
	Unsubscribe(topic Topic, handler func(event interface{}))
}
