package event

import (
	"fmt"
	delegate "github.com/asaskevich/EventBus"
)

type MemoryBus struct {
	delegate delegate.Bus
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		delegate: delegate.New(),
	}
}

func (bus *MemoryBus) Publish(topic Topic, event interface{}) {
	bus.delegate.Publish(topic.id, event)
}

func (bus *MemoryBus) Subscribe(topic Topic, handler func(event interface{})) {
	err := bus.delegate.Subscribe(topic.id, handler)
	if err != nil {
		panic(fmt.Sprintf("Failed to subscribe to topic %s: %s", topic.id, err))
	}
}

func (bus *MemoryBus) Unsubscribe(topic Topic, handler func(event interface{})) {
	err := bus.delegate.Unsubscribe(topic.id, handler)
	if err != nil {
		panic(fmt.Sprintf("Failed to unsunscribe from topic %s: %s", topic.id, err))
	}
}
