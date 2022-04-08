package database

import (
	"errors"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"log"
	"neurobot/infrastructure/database"
)

const testDatabasePath = "neurobot-test.db"

func init() {
	// bump DB log level to fatal errors as triggering an error condition is part of the test
	db.LC().SetLevel(db.LogLevelFatal)

	session := MakeTestDatabaseSession()
	defer session.Close()

	err := database.Migrate(session)
	if err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
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
	settings := sqlite.ConnectionURL{Database: testDatabasePath}

	session, err := sqlite.Open(settings)
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	return session
}
