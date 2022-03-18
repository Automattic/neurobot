package database

import (
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"os"
)

func MakeDatabaseSession() (db.Session, error) {
	databasePath := os.Getenv("DB_FILE")
	settings := sqlite.ConnectionURL{Database: databasePath}

	return sqlite.Open(settings)
}
