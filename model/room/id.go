package room

import (
	"errors"
	"fmt"
	"strings"
)

// ID is a value object for 'room id' in matrix which would store value like !room:matrix.test or #roomalias:matrix.test
type ID interface {
	ID() string
	IsAlias() bool
}

type id struct {
	value string
}

// NewID creates a matrix room 'id' instance from a string value and returns a pointer to it
func NewID(value string) (ID, error) {
	if value == "" {
		return nil, errors.New("room id must not be empty")
	}

	if !strings.HasPrefix(value, "!") && !strings.HasPrefix(value, "#") {
		return nil, fmt.Errorf("room id must start with ! or #, got %s", value)
	}

	if !strings.Contains(value, ":") {
		return nil, fmt.Errorf("room id is missing colon, got %s", value)
	}

	s := strings.Split(value, ":")
	if len(s) == 0 || s[0] == "" || s[1] == "" {
		return nil, fmt.Errorf("room id must have format !foo:example.com or #foo:example.com, got %s", value)
	}

	return &id{value: value}, nil
}

func (id *id) ID() string {
	return id.value
}

func (id *id) IsAlias() bool {
	return strings.HasPrefix(id.value, "#")
}
