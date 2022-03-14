package room

import (
	"testing"
)

func TestEmpty(t *testing.T) {
	_, err := NewId("")
	if err == nil {
		t.Error("must not accept empty string")
	}
}

func TestInvalidStartCharacter(t *testing.T) {
	_, err := NewId("a:matrix.test")
	if err == nil {
		t.Errorf("must start with # or !")
	}
}

func TestInvalid(t *testing.T) {
	_, err := NewId("!room")
	if err == nil {
		t.Error("must have a colon followed by a hostname")
	}

	_, err = NewId("!room:")
	if err == nil {
		t.Error("must have a hostname")
	}
}

func TestValid(t *testing.T) {
	_, err := NewId("!room:matrix.test")
	if err != nil {
		t.Errorf("id is valid so there should be no error, got %s", err)
	}
}

func TestId(t *testing.T) {
	id, _ := NewId("!room:matrix.test")
	if id.Id() != "!room:matrix.test" {
		t.Errorf("invalid id, got %s", id.Id())
	}
}

func TestIsAlias(t *testing.T) {
	id, _ := NewId("#room:matrix.test")

	if !id.IsAlias() {
		t.Error("id should be an alias")
	}
}
