package database

import (
	"database/sql"
	"embed"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/upper/db/v4"
	"net/http"
)

// Embed migrations in the binary.
//go:embed migrations/*.sql
var migrationsFilesystem embed.FS

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

	// Retrieve embedded migrations
	migrations, err := httpfs.New(http.FS(migrationsFilesystem), "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("httpfs", migrations, "sqlite3", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
