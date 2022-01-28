package engine

import (
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/upper/db/v4"
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

type tomlMapping map[string]uint64

func parseTOMLDefs(e *engine) error {
	e.log(fmt.Sprintf("Parsing TOML file at %s", e.workflowsDefTOMLFile))

	var def WorkflowDefintionTOML
	_, err := toml.DecodeFile(e.workflowsDefTOMLFile, &def)
	if err != nil {
		return err
	}

	e.log("\nTOML Defs:")
	for _, w := range def.Workflows {
		e.log(fmt.Sprintf("\n[%s] %s (%s) Active=%t", w.Identifier, w.Name, w.Description, w.Active))
		e.log(fmt.Sprintf("\n >> %s %T %+v", w.Trigger.Variety, w.Trigger.Meta, w.Trigger.Meta))
		for ws, s := range w.Steps {
			e.log(fmt.Sprintf("\n\t[%d] %s (%s) Active=%t", ws, s.Name, s.Description, s.Active))
			e.log(fmt.Sprintf("\n\t >> %s %T %+v\n", s.Variety, s.Meta, s.Meta))
		}
	}
	e.log("\n")

	// Semantic check on data
	if err = runSemanticCheckOnTOML(def); err != nil {
		return err
	}

	// Fetch all DB IDs for workflows that we already have in database
	m, err := getTOMLMapping(e.db)
	if err != nil {
		return err
	}

	// Import data
	for _, w := range def.Workflows {
		id, exist := m[w.Identifier]
		if exist {
			updateTOMLWorkflow(e.db, id, w)
		} else {
			insertTOMLWorkflow(e.db, w)
		}
	}

	return nil
}

func runSemanticCheckOnTOML(def WorkflowDefintionTOML) error {
	// > make sure Identifier is unique for each workflow and based on that realize what inserts/update needs to happen
	uniqueIDs := make(map[string]bool)
	for _, w := range def.Workflows {
		if _, exist := uniqueIDs[w.Identifier]; exist {
			return fmt.Errorf("duplicate workflows defined in TOML with ID:%s", w.Identifier)
		}
		uniqueIDs[w.Identifier] = true // value is irrelevant for us
	}

	return nil
}

func getTOMLMapping(dbs db.Session) (m tomlMapping, err error) {
	// get all workflow meta rows that have toml identifiers saved
	var wfr []WorkflowMetaRow
	res := dbs.Collection("workflow_meta").Find(db.Cond{"key": "toml_identifier"})
	err = res.All(&wfr)
	if err != nil {
		return
	}

	m = make(map[string]uint64)
	for _, row := range wfr {
		m[row.Value] = row.WorkflowID
	}

	return m, nil
}

func insertTOMLWorkflow(dbs db.Session, w WorkflowTOML) error {
	// insert workflow
	iwr, err := dbs.Collection("workflows").Insert(WorkflowRow{
		Name:        w.Name,
		Description: w.Description,
		Active:      boolToInt(w.Active),
	})
	if err != nil {
		return err
	}

	// inserted workflow ID
	wid := uint64(iwr.ID().(int64))

	// insert workflow meta
	insertWorkflowMeta(dbs, wid, "toml_identifier", w.Identifier)
	insertWorkflowMeta(dbs, wid, "workflow_steps_hash", asSha256(w.Steps))

	// insert trigger
	itr, err := dbs.Collection("triggers").Insert(TriggerRow{
		Name:        w.Trigger.Name,
		Description: w.Trigger.Description,
		Variety:     w.Trigger.Variety,
		WorkflowID:  wid,
		Active:      boolToInt(w.Active),
	})
	if err != nil {
		return err
	}

	// inserted trigger ID
	tid := uint64(itr.ID().(int64))

	// insert trigger meta
	for key, value := range w.Trigger.Meta {
		insertTriggerMeta(dbs, tid, key, value)
	}

	// lastly, insert workflow steps
	return insertWFSteps(dbs, wid, w.Steps)
}

func updateTOMLWorkflow(dbs db.Session, id uint64, w WorkflowTOML) error {
	// update workflow basic details
	r := WorkflowRow{}
	res := dbs.Collection("workflows").Find(id)
	res.One(&r)
	r.Name = w.Name
	r.Active = boolToInt(w.Active)
	r.Description = w.Description
	err := res.Update(r)
	if err != nil {
		return err
	}

	// update trigger
	if err := updateTrigger(dbs, id, w.Trigger); err != nil {
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
