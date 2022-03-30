package toml

import (
	"crypto/sha256"
	"fmt"
	"log"
	"neurobot/model/workflow"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/upper/db/v4"
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

	err = runSemanticCheckOnTOML(def)
	if err != nil {
		return fmt.Errorf("semantic checks failed on toml definition: %w", err)
	}

	m, err := wfrepo.GetTOMLMapping()
	if err != nil {
		return fmt.Errorf("toml mapping could not be retrieved: %w", err)
	}

	// Import data
	for _, w := range def.Workflows {
		id, exist := m[w.Identifier]
		if exist {
			err = updateTOMLWorkflow(e.db, id, w)
		} else {
			err = insertTOMLWorkflow(e.db, w)
		}

		if err != nil {
			return err
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

func insertTOMLWorkflow(dbs db.Session, w workflowTOML) error {
	// insert workflow
	iwr, err := dbs.Collection("workflows").Insert(model.Workflow{
		Name:        w.Name,
		Description: w.Description,
		Active:      w.Active,
	})
	if err != nil {
		return err
	}

	// inserted workflow ID
	wid := uint64(iwr.ID().(int64))

	// insert workflow meta
	insertWorkflowMeta(dbs, wid, "toml_identifier", w.Identifier)
	insertWorkflowMeta(dbs, wid, "workflow_steps_hash", asSha256(w.Steps))

	// lastly, insert workflow steps
	return insertWFSteps(dbs, wid, w.Steps)
}

func updateTOMLWorkflow(dbs db.Session, id uint64, w workflowTOML) error {
	// update workflow basic details
	r := model.Workflow{}
	res := dbs.Collection("workflows").Find(id)
	res.One(&r)
	r.Name = w.Name
	r.Active = w.Active
	r.Description = w.Description
	err := res.Update(r)
	if err != nil {
		return err
	}

	// updating workflow steps is a little complicated, allow me to explain
	//
	// first we identify are there any updates to make
	// if not, simply skip
	//
	// if there are updates, are number of steps still same?
	// if there are more steps now, that requires overwrite + insert
	// if there are less steps now, that requires overwrite + delete
	// if there are same steps, that just requires overwrite
	//
	// in addition to that, when a step is changed, their meta needs to be changed as well
	// detecting if just a order has changed, is even more code
	//
	// OR
	//
	// a simpler approach is to just purge all workflow step and step meta rows when there is an update and just insert them fresh
	// this only happens at startup, so isn't really a performance concern, plus keeps the code quite simple
	//
	// code below is of latter approach
	//
	// has workflow steps changed since last time?
	if asSha256(w.Steps) != getWorkflowMeta(dbs, id, "workflow_steps_hash") {
		// delete old data
		if err := deleteAllWFSteps(dbs, id); err != nil {
			return err
		}

		// insert fresh data
		if err := insertWFSteps(dbs, id, w.Steps); err != nil {
			return err
		}

		// update workflow meta
		// Note: "toml_identifier" meta should never be updated
		updateWorkflowMeta(dbs, id, "workflow_steps_hash", asSha256(w.Steps))
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func intSliceToStringSlice(a []uint64) []string {
	b := make([]string, len(a))
	for i, v := range a {
		b[i] = strconv.FormatUint(v, 10)
	}

	return b
}
