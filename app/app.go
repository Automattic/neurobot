package app

import "neurobot/infrastructure/event"

type app struct {
	eventBus event.Bus
}

func NewApp(eventBus event.Bus) *app {
	return &app{
		eventBus: eventBus,
	}
}

func (app app) Run() (err error) {
	// TODO

	// go bus.Subscribe(event.TriggerTopic(), func(event interface{}) {
	//	// do something with the event
	// })

	return err
}
