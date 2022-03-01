package engine

import "testing"

func TestEngineBoots(t *testing.T) {
	engine := NewEngine(
		RunParams{
			Debug:                true,
			IsMatrix:             false,
			Database:             "",
			PortWebhookListener:  "",
			WorkflowsDefTOMLFile: "",
			MatrixHomeServer:     "",
			MatrixUsername:       "",
			MatrixPassword:       "",
		},
	)

	engine.StartUpLite()
}
