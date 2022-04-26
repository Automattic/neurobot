package command

import (
	"fmt"
	"strings"
)

// Command represents a command invoked in a message for a particular bot to react
type Command struct {
	command string
	args    map[string]string
}

func (c *Command) String() string {
	arguments := make([]string, 0, len(c.args))
	for _, arg := range c.args {
		arguments = append(arguments, arg)
	}
	return fmt.Sprintf("!%s %s", c.command, strings.Join(arguments, " "))
}

// NewCommand creates an instance of representation of Command (name + arguments)
func NewCommand(msg string) *Command {
	words := strings.Fields(msg)

	args := make(map[string]string)
	for index, word := range words[1:] {
		argName := fmt.Sprintf("arg%d", index)
		args[argName] = word
	}

	return &Command{
		command: words[0],
		args:    args,
	}
}
