package database

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"os"
	"path/filepath"
	"runtime"
)

// The URL to the directory containing migrations
var migrationsUrl string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDirectory := filepath.Dir(currentFile)

	migrationsUrl = fmt.Sprintf("file://%s/migration/", currentDirectory)
}

func MakeDatabaseSession() (db.Session, error) {
	databasePath := os.Getenv("DB_FILE")
	settings := sqlite.ConnectionURL{Database: databasePath}

	return sqlite.Open(settings)
}

func Migrate(session db.Session) error {
	database, err := sql.Open("sqlite3", session.ConnectionURL().String())
	if err != nil {
		return err
	}
	defer database.Close()

	driver, err := sqlite3.WithInstance(database, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsUrl, "sqlite3", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
