package event

import (
	"fmt"
	delegate "github.com/asaskevich/EventBus"
)

type memoryBus struct {
	delegate delegate.Bus
}

func NewMemoryBus() *memoryBus {
	return &memoryBus{
		delegate: delegate.New(),
	}
}

func (bus *memoryBus) Publish(topic Topic, event interface{}) {
	bus.delegate.Publish(topic.id, event)
}

func (bus *memoryBus) Subscribe(topic Topic, handler func(event interface{})) {
	err := bus.delegate.Subscribe(topic.id, handler)
	if err != nil {
		panic(fmt.Sprintf("Failed to subscribe to topic %s: %s", topic.id, err))
	}
}

func (bus *memoryBus) Unsubscribe(topic Topic, handler func(event interface{})) {
	err := bus.delegate.Unsubscribe(topic.id, handler)
	if err != nil {
		panic(fmt.Sprintf("Failed to unsunscribe from topic %s: %s", topic.id, err))
	}
}
