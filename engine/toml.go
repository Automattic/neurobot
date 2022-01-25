package engine

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/upper/db/v4"
)

type WorkflowDefintionTOML struct {
	Workflows []struct {
		Identifier         string
		Active             bool
		Name               string
		Description        string
		TriggerName        string
		TriggerDescription string
		TriggerVariety     string
		TriggerMeta        map[string]string
		Steps              []struct {
			Active      bool
			Name        string
			Description string
			Variety     string
			Meta        map[string]string
		}
	} `toml:"Workflow"`
}

type tomlMapping map[string]uint64

func parseTOMLDefs(e *engine) error {
	e.log(fmt.Sprintf("Parsing TOML file at %s", e.workflowsDefTOMLFile))

	var def WorkflowDefintionTOML
	_, err := toml.DecodeFile(e.workflowsDefTOMLFile, &def)
	if err != nil {
		return err
	}

	if e.debug {
		e.log("\nTOML Defs:")
		for _, w := range def.Workflows {
			fmt.Printf("\n[%s] %s (%s) Active=%t", w.Identifier, w.Name, w.Description, w.Active)
			fmt.Printf("\n >> %s %T %+v", w.TriggerVariety, w.TriggerMeta, w.TriggerMeta)
			for ws, s := range w.Steps {
				fmt.Printf("\n\t[%d] %s (%s) Active=%t", ws, s.Name, s.Description, s.Active)
				fmt.Printf("\n\t >> %s %T %+v\n", s.Variety, s.Meta, s.Meta)
			}
		}
		e.log("\n")
	}

	// Semantic check on data
	// > make sure ID is unique for each workflow and based on that realize what inserts/update needs to happen
	uniqueIDs := make(map[string]bool)
	for _, w := range def.Workflows {
		if _, exist := uniqueIDs[w.Identifier]; exist {
			log.Fatalf("Duplicate workflows defined with ID:%s", w.Identifier)
		}
		uniqueIDs[w.Identifier] = true // value is irrelevant for us
	}

	// Import data
	// > Fetch all DB IDs for workflows that we have here
	m, err := getTOMLMapping(e.db)
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range def.Workflows {
		id, exist := m[w.Identifier]
		if exist {
			// update workflow basic details
			r := WorkflowRow{}
			res := e.db.Collection("workflows").Find(id)
			res.One(&r)
			r.Name = w.Name
			r.Active = boolToInt(w.Active)
			r.Description = w.Description
			err = res.Update(r)
			if err != nil {
				log.Fatal(err)
			}

			// update trigger basic details
			tr := TriggerRow{}
			res = e.db.Collection("triggers").Find(db.Cond{"workflow_id": id})
			res.One(&tr)

			hasTriggerVarietyChanged := false
			if tr.Variety != w.TriggerVariety {
				hasTriggerVarietyChanged = true
			}

			tr.Name = w.TriggerName
			tr.Description = w.TriggerDescription
			tr.Variety = w.TriggerVariety
			err = res.Update(tr)
			if err != nil {
				log.Fatal(err)
			}

			// update trigger meta
			// delete all trigger meta rows first, if variety has changed
			if hasTriggerVarietyChanged {
				e.db.SQL().Exec(fmt.Sprintf("DELETE from trigger_meta WHERE trigger_id = %d", tr.ID))
			}
			for key, value := range w.TriggerMeta {
				updateTriggerMeta(e.db, tr.ID, key, value)
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
			if asSha256(w.Steps) != getWorkflowMeta(e.db, id, "workflow_steps_hash") {
				// delete old data
				rows := []WFStepRow{}
				res = e.db.Collection("workflow_steps").Find(db.Cond{"workflow_id": id})
				res.All(&rows)

				// find all ids for workflow steps, required to delete meta rows
				var collect []uint64
				for _, row := range rows {
					collect = append(collect, row.ID)
				}

				// delete all workfow step rows
				if err := res.Delete(); err != nil {
					log.Fatal(err)
				}

				// delete all workflow step meta rows
				e.db.SQL().Exec(
					fmt.Sprintf(
						"DELETE from workflow_step_meta WHERE step_id IN (%s)",
						strings.Join(
							intSliceToStringSlice(collect),
							",",
						),
					),
				)

				// insert fresh data
				for i, ws := range w.Steps {
					// insert workflow step
					isr, err := e.db.Collection("workflow_steps").Insert(WFStepRow{
						Name:        ws.Name,
						Description: ws.Description,
						Variety:     ws.Variety,
						WorkflowID:  id,
						SortOrder:   uint64(i),
						Active:      boolToInt(ws.Active),
					})
					if err != nil {
						return err
					}

					// inserted step ID
					sid := uint64(isr.ID().(int64))

					// insert step meta
					for key, value := range ws.Meta {
						insertWFStepMeta(e.db, sid, key, value)
					}
				}

				// update workflow meta
				// Note: "toml_identifier" meta should never be updated
				updateWorkflowMeta(e.db, id, "workflow_steps_hash", asSha256(w.Steps))
			}

		} else {
			// insert workflow
			iwr, err := e.db.Collection("workflows").Insert(WorkflowRow{
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
			insertWorkflowMeta(e.db, wid, "toml_identifier", w.Identifier)
			insertWorkflowMeta(e.db, wid, "workflow_steps_hash", asSha256(w.Steps))

			// insert trigger
			itr, err := e.db.Collection("triggers").Insert(TriggerRow{
				Name:        w.TriggerName,
				Description: w.TriggerDescription,
				Variety:     w.TriggerVariety,
				WorkflowID:  wid,
				Active:      boolToInt(w.Active),
			})
			if err != nil {
				return err
			}

			// inserted trigger ID
			tid := uint64(itr.ID().(int64))

			// insert trigger meta
			for key, value := range w.TriggerMeta {
				insertTriggerMeta(e.db, tid, key, value)
			}

			// insert workflow steps
			for i, ws := range w.Steps {
				// insert workflow step
				isr, err := e.db.Collection("workflow_steps").Insert(WFStepRow{
					Name:        ws.Name,
					Description: ws.Description,
					Variety:     ws.Variety,
					WorkflowID:  wid,
					SortOrder:   uint64(i),
					Active:      boolToInt(ws.Active),
				})
				if err != nil {
					return err
				}

				// inserted step ID
				sid := uint64(isr.ID().(int64))

				// insert step meta
				for key, value := range ws.Meta {
					insertWFStepMeta(e.db, sid, key, value)
				}
			}
		}
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
