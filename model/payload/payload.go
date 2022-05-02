package payload

// Payload represents the data based on which workflows run, passing this data through each workflow step
type Payload struct {
	// what matrix rooms should the message be posted to
	Room string // this might be better of as plural, need to revisit soon
	// what matrix users are relevant for this workflow
	Users []string
	// what message is to be posted in matrix room or elsewhere
	Message string
	// additional contextual parameters
	Context map[string]string
}
