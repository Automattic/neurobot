package app

import "neurobot/infrastructure/event"

// Run contains all our app's entry points, e.g. a CLI, an incoming HTTP request, or an event coming from the event bus.
func Run(bus event.Bus) {
	// TODO

	// go bus.Subscribe(event.TriggerTopic(), func(event interface{}) {
	//	// do something with the event
	// })
}
