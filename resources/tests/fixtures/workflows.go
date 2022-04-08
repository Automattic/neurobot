package fixtures

import (
	"log"
	"neurobot/model/workflow"

	"github.com/upper/db/v4"
)

type workflowMetaRow struct {
	ID         uint64 `db:"id,omitempty"`
	WorkflowID uint64 `db:"workflow_id"`
	Key        string `db:"key"`
	Value      string `db:"value"`
}

func Workflows(session db.Session) map[string]workflow.Workflow {
	fixtures := map[string]workflow.Workflow{
		"QuickStart Demo": {
			ID:          1,
			Name:        "QuickStart Demo",
			Description: "This workflow is meant to show a quick demo",
			Active:      true,
			Identifier:  "QUICKSTART",
		},
		"MVP": {
			ID:          11,
			Name:        "MVP",
			Description: "",
			Active:      true,
			Identifier:  "MVP",
		},
		"Deactivated Workflow": {
			ID:          12,
			Name:        "Deactivated Workflow",
			Description: "",
			Active:      false,
			Identifier:  "DEACTIVATED",
		},
		"Toml imported Workflow": {
			ID:          13,
			Name:        "Toml imported Workflow",
			Description: "",
			Active:      true,
			Identifier:  "TOMLTEST1",
		},
		"Toml imported Workflow 2": {
			ID:          14,
			Name:        "Toml imported Workflow 2",
			Description: "",
			Active:      false,
			Identifier:  "TOMLTEST2",
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
