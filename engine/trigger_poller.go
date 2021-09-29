package engine

import "time"

type polltMeta struct {
	url             string
	endpointType    string
	pollingInterval time.Duration
}

type pollt struct {
	trigger
	polltMeta
}

func (t *pollt) setup() {
	for {
		t.process(nil)
		time.Sleep(t.pollingInterval)
	}
}

func (t *pollt) process(payload interface{}) {
	time.Sleep(2 * time.Second)   // fake processing
	message := "RSS poll results" // fake result

	t.finish(message)
}

func NewRSSPollTrigger(name string, description string, url string, pollingInterval time.Duration, engine *Engine) *pollt {
	return &pollt{
		trigger: trigger{
			variety:     "poller",
			name:        name,
			description: description,
			engine:      engine,
		},
		polltMeta: polltMeta{
			url:             url,
			endpointType:    "rss",
			pollingInterval: pollingInterval,
		},
	}
}

func NewHTTPPollTrigger(name string, description string, url string, pollingInterval time.Duration, engine *Engine) *pollt {
	return &pollt{
		trigger: trigger{
			variety:     "poller",
			name:        name,
			description: description,
			engine:      engine,
		},
		polltMeta: polltMeta{
			url:             url,
			endpointType:    "http",
			pollingInterval: pollingInterval,
		},
	}
}
