package sqlite

import (
	"database/sql"
	"testing"
)

func TestSchemaInit(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !tableExists(t, db, "tasks") {
		t.Fatalf("tasks table should exist")
	}
	if !indexExists(t, db, "idx_tasks_status") {
		t.Fatalf("idx_tasks_status should exist")
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	return db
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name = ?`, name).Scan(&cnt); err != nil {
		t.Fatalf("query table exists: %v", err)
	}
	return cnt > 0
}

func indexExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='index' AND name = ?`, name).Scan(&cnt); err != nil {
		t.Fatalf("query index exists: %v", err)
	}
	return cnt > 0
}
