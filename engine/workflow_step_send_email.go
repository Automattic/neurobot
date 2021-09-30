package engine

import "fmt"

type sendEmailWorkflowStepMeta struct {
	emailAddr string
}

type sendEmailWorkflowStep struct {
	workflowStep
	sendEmailWorkflowStepMeta
}

func (s sendEmailWorkflowStep) run(payload string, e *Engine) string {
	// send email

	// hack: only show decorated log for now
	fmt.Println("====================")
	fmt.Printf("To: %s\n", s.emailAddr)
	fmt.Println(payload)
	fmt.Println("====================")

	return payload
}

func NewSendEmailWorkFlowStep(name string, description string, payload string, email string) *sendEmailWorkflowStep {
	return &sendEmailWorkflowStep{
		workflowStep: workflowStep{
			variety:     "sendEmail",
			name:        name,
			description: description,
			payload:     payload,
		},
		sendEmailWorkflowStepMeta: sendEmailWorkflowStepMeta{
			emailAddr: email,
		},
	}
}
