package engine

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const latestDBVersion = 1

type dbAssist struct {
	debug bool
	db    *sql.DB
}

func (dbA *dbAssist) createTable() (err error) {
	createTableQuery := `CREATE TABLE workflows (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		active integer
	  );
	  CREATE TABLE triggers (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		variety TEXT,
		workflows TEXT
	  );
	  CREATE TABLE workflowsteps (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		variety TEXT,
		workflow integer,
		sortOrder integer
	  );
	  CREATE TABLE options (
		key TEXT NOT NULL PRIMARY KEY,
		value TEXT
	  );`

	_, err = dbA.db.Exec(createTableQuery)
	if err != nil {
		log.Printf("%q: %s\n", err, createTableQuery)
		return
	}

	return
}

func (dbA *dbAssist) manageDBSchema() (err error) {
	if dbA.debug {
		fmt.Println("Checking for DB schema..")
	}
	tableExistQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name='workflows';`
	rows, err := dbA.db.Query(tableExistQuery)
	if err != nil {
		return
	}
	defer rows.Close()

	var tablename string
	for rows.Next() {
		err = rows.Scan(&tablename)
		if err != nil {
			return
		}
		if dbA.debug {
			fmt.Printf("DB Schema exists. tablename: %s\n", tablename)
		}
	}
	err = rows.Err()
	if err != nil {
		return
	}

	if tablename == "" {
		err = dbA.createTable()
		if err != nil {
			return
		}
	}

	// DB version check to run upgradation of schema
	if dbA.debug {
		fmt.Println("Checking DB Version for running schema upgrades..")
	}
	dbVer, err := dbA.getOption("db_ver")
	ver, _ := strconv.ParseInt(dbVer, 10, 64)
	if ver < latestDBVersion {
		return dbA.runDBSchema(ver)
	}

	return
}

func (dbA *dbAssist) runDBSchema(dbVer int64) (err error) {
	if dbA.debug {
		fmt.Printf("Running DB Schema upgrades based on version:%d LatestVersion:%d\n", dbVer, latestDBVersion)
	}

	if dbVer == 0 {
		err = dbA.updateOption("db_ver", "1") // insert
		if err != nil {
			fmt.Printf("DBSchema upgrade failed. dbVer:%d", dbVer)
			return err
		}
	}

	// if dbVer < 2 {
	// 	// DB upgrade schema would come here
	// 	// return if error encountered
	// 	// Update db_ver
	// 	// dbA.updateOption("db_ver", "2")
	// }

	// if dbVer < 3 {
	// 	// DB upgrade schema would come here
	// 	// return if error encountered
	// 	// Update db_ver
	// 	// dbA.updateOption("db_ver", "3")
	// }

	return
}

func (dbA *dbAssist) getOption(key string) (value string, err error) {
	q := fmt.Sprintf("select value from options where key = '%s';", key)
	rows, err := dbA.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&value)
		if err != nil {
			return
		}
		if dbA.debug {
			fmt.Printf("getOption() %s : %s\n", key, value)
		}
	}
	err = rows.Err()
	if err != nil {
		return
	}

	return
}

func (dbA *dbAssist) updateOption(key string, value string) (err error) {
	v, err := dbA.getOption(key)
	if err != nil {
		return
	}

	if v == value {
		return
	}

	// Insert?
	if v == "" {
		q := "INSERT INTO options (key,value) VALUES (?, ?);"
		var statement *sql.Stmt
		statement, err = dbA.db.Prepare(q)
		if err != nil {
			return
		}
		_, err = statement.Exec(key, value)
		return
	}

	// Update
	q := "UPDATE options SET value = ? WHERE key = 'db_ver';"
	statement, err := dbA.db.Prepare(q)
	if err != nil {
		return
	}
	_, err = statement.Exec(value)
	return
}

func NewDBAssist(db *sql.DB, debug bool) *dbAssist {
	return &dbAssist{
		db:    db,
		debug: debug,
	}
}
