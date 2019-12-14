package db

import (
	"database/sql"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// Connect ...
func Connect(driver string, datasource string, migrations ...Migration) (*DB, error) {
	db, err := sql.Open(driver, datasource)
	if err != nil {
		return nil, err
	}

	switch driver {
	case "mysql":
		db.SetMaxIdleConns(0)
	}
	if err := pingDatabase(db); err != nil {
		return nil, err
	}
	if err := setupDatabase(db, driver, migrations...); err != nil {
		return nil, err
	}

	var engine Driver
	var locker Locker

	switch driver {
	case "mysql":
		engine = MySQL
		locker = &nopLocker{}
	case "postgres":
		engine = Postgres
		locker = &nopLocker{}
	default:
		engine = Sqlite
		locker = &sync.RWMutex{}
	}

	return &DB{
		conn:   sqlx.NewDb(db, driver),
		lock:   locker,
		driver: engine,
	}, nil
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the databsae by performing automated
// database migration steps.
func setupDatabase(db *sql.DB, driver string, migrations ...Migration) error {
	for _, mig := range migrations {
		if mig.Driver == driver {
			return mig.Migrator.Migrate(db)
		}
	}
	return nil
}
