package database

import (
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func MakeDatabaseSession(databasePath string) (db.Session, error) {
	settings := sqlite.ConnectionURL{Database: databasePath}

	db.LC().SetLevel(db.LogLevelError)
	return sqlite.Open(settings)
}
