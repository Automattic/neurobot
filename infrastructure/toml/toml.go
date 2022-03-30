package toml

import (
	"fmt"
	"log"
	"neurobot/model/workflow"
	"neurobot/model/workflowstep"
	"strconv"

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
	Trigger     workflowTriggerTOML
	Steps       []workflowStepTOML `toml:"Step"`
}

type workflowTriggerTOML struct {
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

type workflowStepTOML struct {
	Active      bool
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

// Import accepts a workflow repository where workflows are imported from the provided toml file
func Import(wfrepo workflow.Repository, tomlFilePath string) (err error) {
	def, err := parse(tomlFilePath)
	if err != nil {
		return fmt.Errorf("error while parsing toml file: %w", err)
	}

	workflows, err := prepare(def, wfrepo)
	if err != nil {
		return
	}

	for _, w := range workflows {
		if err = wfrepo.Save(w); err != nil {
			return
		}
	}

	return
}

func parse(tomlFilePath string) (def workflowDefintionTOML, err error) {
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

func prepare(def workflowDefintionTOML, wfrepo workflow.Repository) (prepared []*workflow.Workflow, err error) {
	for _, workflow := range def.Workflows {
		w, err := wfrepo.FindByIdentifier(workflow.Identifier)
		if err != nil {
			return nil, err
		}

		w.Name = workflow.Name
		w.Description = workflow.Description
		w.Active = workflow.Active
		for index, s := range workflow.Steps {
			w.Steps[index] = workflowstep.WorkflowStep{
				Active:      boolToInt(s.Active),
				Name:        s.Name,
				Description: s.Description,
				Variety:     s.Variety,
				Meta:        s.Meta,
			}
		}

		prepared = append(prepared, &w)
	}

	return
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intSliceToStringSlice(a []uint64) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.FormatUint(v, 10)
	}

	return b
}
