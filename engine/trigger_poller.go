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
