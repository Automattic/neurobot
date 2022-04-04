package toml

import (
	"fmt"
	"neurobot/model/workflow"
	"neurobot/model/workflowstep"

	"github.com/BurntSushi/toml"
)

type workflowDefintionTOML struct {
	Workflows []workflowTOML `toml:"Workflow"`
}

type workflowTOML struct {
	Identifier  string
	Active      bool
	Name        string
	Description string
	Steps       []workflowStepTOML `toml:"Step"`
}

type workflowStepTOML struct {
	Active      bool
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

// Import accepts a workflow repository where workflows are to be imported from the provided toml file
func Import(tomlFilePath string, wfRepo workflow.Repository, wfsRepo workflowstep.Repository) (err error) {
	workflowDefs, err := parse(tomlFilePath)
	if err != nil {
		return fmt.Errorf("error while parsing toml file: %w", err)
	}

	for _, def := range workflowDefs.Workflows {
		workflow, workflowSteps, err := prepare(def, wfRepo, wfsRepo)
		if err != nil {
			return fmt.Errorf("error while preparing toml def for import: %w", err)
		}

		if err = wfRepo.Save(&workflow); err != nil {
			return err
		}

		// remove all workflow steps for this workflow before freshly insert all workflow steps data (including step meta)
		if err = wfsRepo.RemoveByWorkflowID(workflow.ID); err != nil {
			return err
		}

		for _, step := range workflowSteps {
			// now that we surely have the workflow ID, populate that in step
			step.WorkflowID = workflow.ID
			if err = wfsRepo.Save(&step); err != nil {
				return err
			}
		}
	}

	return
}

func parse(tomlFilePath string) (def workflowDefintionTOML, err error) {
	_, err = toml.DecodeFile(tomlFilePath, &def)
	if err != nil {
		return
	}

	err = runSemanticCheckOnTOML(def)
	if err != nil {
		return def, fmt.Errorf("semantic checks failed on toml definition: %w", err)
	}

	return
}

func runSemanticCheckOnTOML(def workflowDefintionTOML) error {
	// > make sure Identifier is unique for each workflow and based on that realize what inserts/update needs to happen
	// > make sure workflow has atleast a trigger and atleast a workflow step inside of it
	uniqueIDs := make(map[string]bool)
	for _, w := range def.Workflows {
		if _, exist := uniqueIDs[w.Identifier]; exist {
			return fmt.Errorf("duplicate workflows defined in TOML with ID:%s", w.Identifier)
		}
		uniqueIDs[w.Identifier] = true // value is irrelevant for us

		// no workflow steps defined?
		if len(w.Steps) == 0 {
			return fmt.Errorf("no workflow steps defined for workflow in TOML with ID:%s", w.Identifier)
		}
	}

	return nil
}

// Prepares a workflow struct and an array of workflow steps struct from a TOML definition of a single workflow
func prepare(def workflowTOML, wfRepo workflow.Repository, wfsRepo workflowstep.Repository) (w workflow.Workflow, steps []workflowstep.WorkflowStep, err error) {
	w, _ = wfRepo.FindByIdentifier(def.Identifier)

	w.Identifier = def.Identifier
	w.Name = def.Name
	w.Description = def.Description
	w.Active = def.Active

	for _, step := range def.Steps {
		s := workflowstep.WorkflowStep{
			Active:      step.Active,
			Name:        step.Name,
			Description: step.Description,
			Variety:     step.Variety,
			WorkflowID:  w.ID,
			Meta:        step.Meta,
		}

		steps = append(steps, s)
	}

	return
}
