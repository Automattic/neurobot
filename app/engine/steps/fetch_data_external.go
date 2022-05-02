package steps

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"neurobot/model/payload"
)

type fetchDataExternalMeta struct {
	url string
}

type fetchDataExternalWorkflowStepRunner struct {
	fetchDataExternalMeta
}

func (runner *fetchDataExternalWorkflowStepRunner) Run(p *payload.Payload) error {
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
func NewFetchDataExternalRunner(meta map[string]string) *fetchDataExternalWorkflowStepRunner {
	var stepMeta fetchDataExternalMeta
	stepMeta.url, _ = meta["url"]
	return &fetchDataExternalWorkflowStepRunner{stepMeta}
}
