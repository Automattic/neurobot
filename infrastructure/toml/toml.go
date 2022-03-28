package toml

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type WorkflowDefintionTOML struct {
	Workflows []WorkflowTOML `toml:"Workflow"`
}

type WorkflowTOML struct {
	Identifier  string
	Active      bool
	Name        string
	Description string
	Trigger     WorkflowTriggerTOML
	Steps       []WorkflowStepTOML `toml:"Step"`
}

type WorkflowTriggerTOML struct {
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

type WorkflowStepTOML struct {
	Active      bool
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

func Parse(tomlFilePath string) (def WorkflowDefintionTOML, err error) {
	_, err = toml.DecodeFile(tomlFilePath, &def)
	if err != nil {
		return
	}

	log.Println("\nTOML Defs:")
	for _, w := range def.Workflows {
		log.Printf("\n[%s] %s (%s) Active=%t", w.Identifier, w.Name, w.Description, w.Active)
		log.Printf("\n >> %s %T %+v", w.Trigger.Variety, w.Trigger.Meta, w.Trigger.Meta)
		for ws, s := range w.Steps {
			log.Printf("\n\t[%d] %s (%s) Active=%t", ws, s.Name, s.Description, s.Active)
			log.Printf("\n\t >> %s %T %+v\n", s.Variety, s.Meta, s.Meta)
		}
	}
	log.Println("\n---TOML---")

	return
}

func RunSemanticCheckOnTOML(def WorkflowDefintionTOML) error {
	// > make sure Identifier is unique for each workflow and based on that realize what inserts/update needs to happen
	// > make sure workflow has atleast a trigger and atleast a workflow step inside of it
	uniqueIDs := make(map[string]bool)
	for _, w := range def.Workflows {
		if _, exist := uniqueIDs[w.Identifier]; exist {
			return fmt.Errorf("duplicate workflows defined in TOML with ID:%s", w.Identifier)
		}
		uniqueIDs[w.Identifier] = true // value is irrelevant for us

		// no trigger defined?
		if w.Trigger.Name == "" || w.Trigger.Description == "" || w.Trigger.Variety == "" {
			return fmt.Errorf("no trigger defined for workflow in TOML with ID:%s", w.Identifier)
		}

		// no workflow steps defined?
		if len(w.Steps) == 0 {
			return fmt.Errorf("no workflow steps defined for workflow in TOML with ID:%s", w.Identifier)
		}
	}

	return nil
}
