package database

import (
	"github.com/apex/log"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

func MakeDatabaseSession(databasePath string) db.Session {
	logger := log.WithFields(log.Fields{
		"path": databasePath,
	})

	settings := sqlite.ConnectionURL{Database: databasePath}
	db.LC().SetLevel(db.LogLevelError)

	session, err := sqlite.Open(settings)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	err = Migrate(session)
	if err != nil {
		logger.WithError(err).Fatal("Failed to migrate database")
	}

	return session
}
