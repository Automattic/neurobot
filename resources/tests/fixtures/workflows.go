package fixtures

import (
	"github.com/upper/db/v4"
	"log"
	"neurobot/model/workflow"
)

func Workflows(session db.Session) map[string]workflow.Workflow {
	// Make sure there are no workflows configured elsewhere than these fixtures.
	// TODO: Currently migrations insert the "QuickStart Demo" workflow in the workflows table.
	//       Once that is no longer the case, this truncate can be removed.
	err := session.Collection("workflows").Truncate()
	if err != nil {
		log.Fatalf("Failed to truncate workflows table: %s", err)
	}

	fixtures := map[string]workflow.Workflow{
		"QuickStart Demo": {
			ID:          1,
			Name:        "QuickStart Demo",
			Description: "This workflow is meant to show a quick demo",
			Active:      true,
		},
		"MVP": {
			ID:          11,
			Name:        "MVP",
			Description: "",
			Active:      true,
		},
		"Deactivated Workflow": {
			ID:          12,
			Name:        "Deactivated Workflow",
			Description: "",
			Active:      false,
		},
		"Toml imported Workflow": {
			ID:          13,
			Name:        "Toml imported Workflow",
			Description: "",
			Active:      true,
		},
		"Toml imported Workflow 2": {
			ID:          14,
			Name:        "Toml imported Workflow 2",
			Description: "",
			Active:      false,
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
