package engine

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type WorkflowDefintionTOML struct {
	Workflows []struct {
		ID          uint
		Active      bool
		Name        string
		Description string
		TriggerType string
		TriggerMeta toml.Primitive
		Steps       []struct {
			Active      bool
			Name        string
			Description string
			Type        string
			Meta        toml.Primitive
		}
	} `toml:"Workflow"`
}

func parseTOMLDefs(e *engine) {
	e.log(fmt.Sprintf("Parsing TOML file at %s", e.workflowsDefTOMLFile))

	var def WorkflowDefintionTOML
	md, err := toml.DecodeFile(e.workflowsDefTOMLFile, &def)
	if err != nil {
		log.Fatal(err)
	}

	if e.debug {
		e.log("\n\nTOML Defs (without primitive decoding):\n")
		fmt.Println(def.Workflows)
		for _, w := range def.Workflows {
			fmt.Printf("\n[%d] %s (%s) Active=%t", w.ID, w.Name, w.Description, w.Active)
			fmt.Printf("\n >> %s %T %+v\n", w.TriggerType, w.TriggerMeta, w.TriggerMeta)
			fmt.Printf("\n%+v", w.Steps)
		}
	}

	for ww, w := range def.Workflows {

		// handle trigger meta
		var triggerMeta interface{}
		switch w.TriggerType {
		case "webhook":
			triggerMeta = new(webhooktMeta)
		case "poll":
			triggerMeta = new(polltMeta)
		default:
			log.Fatalf("Unidentified trigger type '%s' encountered for workflow id:%d while parsing TOML file", w.TriggerType, w.ID)
		}

		if err := md.PrimitiveDecode(def.Workflows[ww].TriggerMeta, triggerMeta); err != nil {
			log.Fatalf("Unexpected meta values for trigger found for workflow id:%d while parsing TOML file", w.ID)
		}

		// handle steps meta
		for ws, s := range w.Steps {

			var stepMeta interface{}
			skipPrimitiveDecoding := false // some worksteps don't have any meta values

			switch s.Type {
			case "postMatrixMessage":
				stepMeta = new(postMessageMatrixWorkflowStepMeta)
			case "stdout":
				skipPrimitiveDecoding = true // no meta for "stdout" workflowstep
			default:
				log.Fatalf("Unidentified workflowstep type '%s' encountered for workflow id:%d while parsing TOML file", s.Type, w.ID)
			}

			if !skipPrimitiveDecoding {
				if err := md.PrimitiveDecode(def.Workflows[ww].Steps[ws].Meta, stepMeta); err != nil {
					log.Fatalf("Unexpected meta values for workflowstep found for workflow id:%d while parsing TOML file", w.ID)
				}
			}
		}
	}

	if e.debug {
		e.log("\n\nTOML Defs:\n")
		fmt.Println(def.Workflows)
		for _, w := range def.Workflows {
			fmt.Printf("\n[%d] %s (%s) Active=%t", w.ID, w.Name, w.Description, w.Active)
			fmt.Printf("\n%s %T %+v\n", w.TriggerType, w.TriggerMeta, w.TriggerMeta)
			fmt.Printf("\n%+v", w.Steps)
		}
	}

	// fmt.Println("%%%%%%%%%%%%%")
	// fmt.Println(def.Workflows[1].Steps[0].Meta.messagePrefix)
}
