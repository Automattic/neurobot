package app

import (
	"neurobot/engine"
	"neurobot/infrastructure/event"
	"neurobot/infrastructure/http"
	b "neurobot/model/bot"
	w "neurobot/model/workflow"
)

type app struct {
	engine             engine.Engine
	eventBus           event.Bus
	botRepository      b.Repository
	workflowRepository w.Repository
	webhookListener    *http.Server
}

func NewApp(
	engine engine.Engine,
	eventBus event.Bus,
	botRepository b.Repository,
	workflowRepository w.Repository,
	webhookListener *http.Server,
) *app {
	return &app{
		engine:             engine,
		eventBus:           eventBus,
		botRepository:      botRepository,
		workflowRepository: workflowRepository,
		webhookListener:    webhookListener,
	}
}

func (app app) Run() (err error) {
	// TODO

	// go bus.Subscribe(event.TriggerTopic(), func(event interface{}) {
	//	// do something with the event
	// })

	return err
}
