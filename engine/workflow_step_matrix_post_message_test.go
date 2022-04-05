package engine

import (
	"testing"

	"maunium.net/go/mautrix"
)

func TestGenericGetMatrixClient(t *testing.T) {
	tables := []struct {
		homeserver string
		isErr      bool
	}{
		{
			homeserver: "https://example.org",
			isErr:      false,
		},
		{
			homeserver: "example.org",
			isErr:      false,
		},
		{
			homeserver: " http://foo.com",
			isErr:      true,
		},
	}
	for _, table := range tables {
		c, err := getMatrixClient(table.homeserver)
		if err != nil {
			if !table.isErr {
				t.Error("error thrown when it shouldn't have")
			}

		} else {
			if table.isErr {
				t.Error("error not thrown when it should have")
			}

			if _, ok := c.(*mautrix.Client); !ok {
				t.Error("mautrix client wasn't returned")
			}
		}
	}
}
