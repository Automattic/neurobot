package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"neurobot/infrastructure/event"
	"time"
)

type poller struct {
	interval time.Duration
	url      *url.URL
	eventBus event.Bus
}

func NewHttpPoller(duration string, url *url.URL, eventBus event.Bus) *poller {
	interval, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("Failed to parse duration %s, defaulting to 1 minute", duration)
		interval, _ = time.ParseDuration("1m")
	}

	return &poller{
		interval: interval,
		url:      url,
		eventBus: eventBus,
	}
}

func (poller *poller) Run() {
	for {
		time.Sleep(poller.interval)
		poller.poll()
	}
}

func (poller *poller) poll() {
	response, err := http.Get(poller.url.String())
	if err != nil {
		log.Printf("Failed to poll %s: %s", poller.url.String(), err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to parse response from url %s: %s", poller.url.String(), err)
	}

	// TODO: Remove this once the body variable is being used
	log.Printf(string(body))

	// TODO: create a trigger, then publish it
	// http.eventBus.Publish(event.TriggerTopic(), trigger)
}
