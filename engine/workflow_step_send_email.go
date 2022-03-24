package engine

import "fmt"

type sendEmailWorkflowPayload struct {
	message string
}

type sendEmailWorkflowStepMeta struct {
	emailAddr string
}

type sendEmailWorkflowStep struct {
	workflowStep
	sendEmailWorkflowStepMeta
}

func (s sendEmailWorkflowStep) run(p map[string]string, e *engine) (map[string]string, error) {
	// @TODO: send email
	// hack: only show decorated log for now
	fmt.Println("====================")
	fmt.Printf("To: %s\n", s.emailAddr)
	fmt.Println(p["Message"])
	fmt.Println("====================")

	return p, nil
}
