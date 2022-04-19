package database

import (
	"errors"
	"log"
	"neurobot/infrastructure/database"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func init() {
	settings := sqlite.ConnectionURL{
		Database: ":memory:",
		Options: map[string]string{
			"mode":  "memory",
			"cache": "shared",
		},
	}

	session, err := sqlite.Open(settings)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %s", err)
	}

	err = database.Migrate(session)
	if err != nil {
		log.Fatalf("Failed to run upgrade/migration scripts on test database: %s", err)
	}
}

// Test is a "wrapper" for tests that interact with the Database.
// It wraps the test in a transaction, and rolls it back automatically,
// since in tests, we never want to commit the transaction.
//
// Example usage:
//
// func TestFoo(t *testing.T) {
//     database.Test(func(session db.Session) {
//         // test something
//     })
// }
func Test(fn func(session db.Session)) {
	session := MakeTestDatabaseSession()
	defer session.Close()
	_ = session.Tx(func(session db.Session) error {
		fn(session)

		// Returning an error results in the transaction being rolled back.
		return errors.New("rollback")
	})
}

func MakeTestDatabaseSession() db.Session {
	settings := sqlite.ConnectionURL{
		Database: ":memory:",
		Options: map[string]string{
			"mode":  "memory",
			"cache": "shared",
		},
	}

	session, err := sqlite.Open(settings)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %s", err)
	}

	return session
}
