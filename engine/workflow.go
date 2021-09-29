package engine

import (
	"fmt"
)

type workflow struct {
	id          uint64
	name        string
	description string
	payload     interface{}
	steps       []WorkflowStep
}

func (w *workflow) addWorkflowStep(s WorkflowStep) {
	w.steps = append(w.steps, s)
}

func (w *workflow) run(payload string) {
	fmt.Printf("\nRunning workflow #%d\n", w.id)
	// loop through all the steps inside of this workflow
	for _, s := range w.steps {
		w.payload = s.run(payload) // overwrite payload with each step execution and keep on passing this payload to each step
	}
}

type WorkflowStep interface {
	run(string) string // accepts payload from workflow and returns after modification (if desired)
}

type workflowStep struct {
	variety     string
	name        string
	description string
	payload     string
}

type sendEmailWorkflowStepMeta struct {
	emailAddr string
}

type sendEmailWorkflowStep struct {
	workflowStep
	sendEmailWorkflowStepMeta
}

func (s sendEmailWorkflowStep) run(m string) string {
	// send email

	// hack: only show decorated log for now
	fmt.Println("====================")
	fmt.Printf("To: %s\n", s.emailAddr)
	fmt.Println(m)
	fmt.Println("====================")

	return m
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
