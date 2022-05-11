package command

// CommandsHandler defines the interface for command handler with which bot registry communicates with the app
type CommandsHandler interface {
	Send(*Command)
	Run() chan *Command
}

type commandsHandler struct {
	commandChan chan *Command
}

func (h *commandsHandler) Send(c *Command) {
	h.commandChan <- c
}

func (h *commandsHandler) Run() chan *Command {
	return h.commandChan
}

// NewCommandsHandler returns an instance of commands handler for bot registry to notify when a command is sighted in a message and app to act on that notification to run that command
func NewCommandsHandler() CommandsHandler {
	return &commandsHandler{
		commandChan: make(chan *Command),
	}
}
