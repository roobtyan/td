package usecase

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"td/internal/repo/sqlite"
)

func TestAddFromClipboardRule(t *testing.T) {
	db := sqliteOpenTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := sqlite.NewTaskRepository(db)
	uc := AddFromClipboardUseCase{
		Repo: repo,
	}

	task, err := uc.AddFromClipboard(context.Background(), "Buy milk\nhttps://example.com", false)
	if err != nil {
		t.Fatalf("add from clipboard: %v", err)
	}
	if task.Title == "" {
		t.Fatalf("task title should not be empty")
	}
}

func sqliteOpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	return db
}
