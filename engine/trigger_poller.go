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
		t.process(payloadData{})
		time.Sleep(t.pollingInterval)
	}
}

func (t *pollt) process(p payloadData) {
	time.Sleep(2 * time.Second)    // fake processing
	p.Message = "RSS poll results" // fake result

	t.finish(p)
}
