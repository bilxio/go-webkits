package db

import "database/sql"

// Migrator ...
type Migrator interface {
	Migrate(*sql.DB) error
}

// MigratorFunc represnts a migration for a database
type MigratorFunc func(*sql.DB) error

// Migrate ...
func (m MigratorFunc) Migrate(db *sql.DB) error {
	return m(db)
}

// Migration needs a drvier name and its migration func
type Migration struct {
	Driver   string
	Migrator Migrator
}
