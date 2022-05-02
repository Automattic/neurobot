package steps

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"neurobot/model/payload"

	"github.com/apex/log"
)

type fetchDataExternalMeta struct {
	url string
}

type fetchDataExternalWorkflowStepRunner struct {
	eid string
	fetchDataExternalMeta
}

func (runner *fetchDataExternalWorkflowStepRunner) Run(p *payload.Payload) error {
	log.Log.WithFields(log.Fields{
		"executionID":  runner.eid,
		"workflowStep": "fetchDataExternal",
	}).Info("running workflow step")

	j, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", runner.url, bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth", "secret") // @TODO move this to config

	client := &http.Client{}
	client.Timeout = time.Second * 120

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err
	}

	// overwrite payload and return
	json.NewDecoder(resp.Body).Decode(&p)
	return nil
}

// NewFetchDataExternalRunner returns an instance of worklow step for fetching data from an external source
func NewFetchDataExternalRunner(eid string, meta map[string]string) *fetchDataExternalWorkflowStepRunner {
	var stepMeta fetchDataExternalMeta
	stepMeta.url, _ = meta["url"]
	return &fetchDataExternalWorkflowStepRunner{
		eid:                   eid,
		fetchDataExternalMeta: stepMeta,
	}
}
