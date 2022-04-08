package fixtures

import (
	"log"
	"neurobot/model/workflowstep"

	"github.com/upper/db/v4"
)

type workflowStepMetaRow struct {
	ID     uint64 `db:"id,omitempty"`
	StepID uint64 `db:"step_id"`
	Key    string `db:"key"`
	Value  string `db:"value"`
}

func WorkflowSteps(session db.Session) map[string]workflowstep.WorkflowStep {
	fixtures := map[string]workflowstep.WorkflowStep{
		"PostMessage1": {
			ID:          1,
			Name:        "Post Message to Matrix Room 1",
			Description: "Description",
			Variety:     "postMatrixMessage",
			WorkflowID:  1,
			SortOrder:   0,
			Active:      true,
			Meta: map[string]string{
				"matrixRoom":    "#orbit",
				"messagePrefix": "[Alert]",
			},
		},
		"PostMessage2": {
			ID:          2,
			Name:        "Post Message to Matrix Room 2",
			Description: "Some Description",
			Variety:     "postMatrixMessage",
			WorkflowID:  2,
			SortOrder:   1,
			Active:      false,
			Meta: map[string]string{
				"matrixRoom":    "#neso",
				"messagePrefix": "[FYI]",
			},
		},
		"PostMessage3": {
			ID:          3,
			Name:        "Post Message to Matrix Room 2",
			Description: "Some Description",
			Variety:     "postMatrixMessage",
			WorkflowID:  1,
			SortOrder:   1,
			Active:      false,
			Meta: map[string]string{
				"matrixRoom":    "#neso",
				"messagePrefix": "[FYI]",
			},
		},
	}

	for _, fixture := range fixtures {
		_, err := session.Collection("workflow_steps").Insert(fixture)
		if err != nil {
			log.Fatalf("Failed to insert fixtures for workflow steps: %s", err)
		}

		for k, v := range fixture.Meta {
			_, err = session.Collection("workflow_step_meta").Insert(workflowStepMetaRow{
				StepID: fixture.ID,
				Key:    k,
				Value:  v,
			})
			if err != nil {
				log.Fatalf("Failed to insert fixtures for workflow step meta: %s", err)
			}
		}

	}

	return fixtures
}
