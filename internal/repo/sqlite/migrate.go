package sqlite

import (
	"database/sql"
	_ "embed"
)

//go:embed migrations/0001_init.sql
var migration0001 string

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(migration0001); err != nil {
		return err
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO schema_migrations(version) VALUES (1)`); err != nil {
		return err
	}
	return nil
}
