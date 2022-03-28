package toml

import (
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
