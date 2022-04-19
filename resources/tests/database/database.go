package database

import (
	"errors"
	"neurobot/infrastructure/database"

	"github.com/apex/log"
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
		log.WithError(err).Fatal("Failed to connect to test database")
	}

	err = database.Migrate(session)
	if err != nil {
		log.WithError(err).Fatal("Failed to migrate test database")
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

// MakeTestDatabaseSession creates an in-memory SQLite database to use in running tests
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
		log.WithError(err).Fatal("Failed to connect to test database")
	}

	return session
}
