package commands

// Command defines the interface for each command that is to be defined
type Command interface {
	// Valid method returns true/false based on whether it has valid data for execution
	Valid() bool
	// UsageHints method simply shows an example for command invokation
	UsageHints() string
	// Returns the payload prepared internally, since payload comes into existence once command triggers
	WorkflowPayload() map[string]string
}
