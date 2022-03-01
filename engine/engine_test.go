package engine

import "testing"

func TestEngineBoots(t *testing.T) {
	engine := NewEngine(
		RunParams{
			Debug:                true,
			IsMatrix:             false,
			Database:             "../neurobot-test.db",
			PortWebhookListener:  "",
			WorkflowsDefTOMLFile: "",
			MatrixHomeServer:     "",
			MatrixUsername:       "",
			MatrixPassword:       "",
		},
	)

	engine.StartUpLite()
}