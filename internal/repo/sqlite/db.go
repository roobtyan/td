package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err = Migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
