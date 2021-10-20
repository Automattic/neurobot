package engine

import "fmt"

type sendEmailWorkflowStepMeta struct {
	emailAddr string
}

type sendEmailWorkflowStep struct {
	workflowStep
	sendEmailWorkflowStepMeta
}

func (s sendEmailWorkflowStep) run(payload string, e *engine) string {
	// send email

	// hack: only show decorated log for now
	fmt.Println("====================")
	fmt.Printf("To: %s\n", s.emailAddr)
	fmt.Println(payload)
	fmt.Println("====================")

	return payload
}
