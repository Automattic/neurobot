package room

import (
	"errors"
	"fmt"
	"strings"
)

type Id interface {
	Id() string
	IsAlias() bool
}

type id struct {
	value string
}

func NewId(value string) (Id, error) {
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

func (id *id) Id() string {
	return id.value
}

func (id *id) IsAlias() bool {
	return strings.HasPrefix(id.value, "#")
}
