package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"td/internal/domain"
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

func TestAddFromClipboardAIShouldSetDueAndTodoStatus(t *testing.T) {
	db := sqliteOpenTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := sqlite.NewTaskRepository(db)
	uc := AddFromClipboardUseCase{
		Repo: repo,
		AIParser: &AIParseTaskUseCase{
			Provider: fakeParseProvider{
				raw: `{"title":"完成在线推理功能","due":"2026-02-25 13:00","priority":"P2"}`,
			},
		},
	}

	task, err := uc.AddFromClipboard(context.Background(), "明天下午一点完成在线推理功能", true)
	if err != nil {
		t.Fatalf("add from clipboard ai: %v", err)
	}
	if task.Status != domain.StatusTodo {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusTodo)
	}
	if task.DueAt == nil {
		t.Fatalf("due should not be nil")
	}
	wantDue, err := time.ParseInLocation("2006-01-02 15:04", "2026-02-25 13:00", time.Local)
	if err != nil {
		t.Fatalf("parse want due: %v", err)
	}
	if !task.DueAt.Equal(wantDue) {
		t.Fatalf("due = %s, want %s", task.DueAt.Format("2006-01-02 15:04"), wantDue.Format("2006-01-02 15:04"))
	}
}

func TestParseInputShouldReturnAISourceAndProject(t *testing.T) {
	uc := AddFromClipboardUseCase{
		AIParser: &AIParseTaskUseCase{
			Provider: fakeParseProvider{
				raw: `{"title":"完成在线推理功能","project":"PP Polaris","priority":"P1"}`,
			},
		},
	}

	parsed, source, err := uc.ParseInput(context.Background(), "明天下午一点完成在线推理功能", true)
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}
	if source != "ai" {
		t.Fatalf("source = %q, want %q", source, "ai")
	}
	if parsed.Project != "PP Polaris" {
		t.Fatalf("project = %q, want %q", parsed.Project, "PP Polaris")
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
