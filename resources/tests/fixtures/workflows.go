package fixtures

import (
	"github.com/upper/db/v4"
	"log"
	"neurobot/model/workflow"
)

func Workflows(session db.Session) map[string]workflow.Workflow {
	fixtures := map[string]workflow.Workflow{
		"MVP": {
			ID:          11,
			Name:        "MVP",
			Description: "",
			Active:      1,
		},
		"Deactivated Workflow": {
			ID:          12,
			Name:        "Deactivated Workflow",
			Description: "",
			Active:      0,
		},
		"Toml imported Workflow": {
			ID:          13,
			Name:        "Toml imported Workflow",
			Description: "",
			Active:      1,
		},
		"Toml imported Workflow 2": {
			ID:          14,
			Name:        "Toml imported Workflow 2",
			Description: "",
			Active:      0,
		},
	}

	for _, fixture := range fixtures {
		_, err := session.Collection("workflows").Insert(fixture)
		if err != nil {
			log.Fatalf("Failed to insert fixtures for workflows: %s", err)
		}
	}

	return fixtures
}
