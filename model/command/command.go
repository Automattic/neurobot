package command

import (
	"fmt"
	"strings"
)

// Command represents a command invoked in a message for a particular bot to react
type Command struct {
	Name string
	Args map[string]string
}

func (c *Command) String() string {
	arguments := make([]string, 0, len(c.Args))
	for _, arg := range c.Args {
		arguments = append(arguments, arg)
	}
	return fmt.Sprintf("!%s %s", c.Name, strings.Join(arguments, " "))
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
		Name: words[0],
		Args: args,
	}
}
