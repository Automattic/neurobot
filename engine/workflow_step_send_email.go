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

func (s sendEmailWorkflowStep) run(payload interface{}, e *engine) (interface{}, error) {
	p := payload.(sendEmailWorkflowPayload)
	// send email

	// hack: only show decorated log for now
	fmt.Println("====================")
	fmt.Printf("To: %s\n", s.emailAddr)
	fmt.Println(p.message)
	fmt.Println("====================")

	return payload, nil
}
