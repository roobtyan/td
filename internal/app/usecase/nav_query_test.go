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

func TestTodayView(t *testing.T) {
	db := openNavTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	seedNavTask(t, db, "doing-a", domain.StatusDoing, "", nil, nil)
	seedNavTask(t, db, "todo-overdue", domain.StatusTodo, "", ptrTime(now.Add(-26*time.Hour)), nil)
	seedNavTask(t, db, "todo-today", domain.StatusTodo, "", ptrTime(now.Add(2*time.Hour)), nil)
	seedNavTask(t, db, "todo-future", domain.StatusTodo, "", ptrTime(now.Add(30*time.Hour)), nil)
	seedNavTask(t, db, "inbox-today", domain.StatusInbox, "", ptrTime(now.Add(1*time.Hour)), nil)
	seedNavTask(t, db, "done-overdue", domain.StatusDone, "", ptrTime(now.Add(-2*time.Hour)), ptrTime(now.Add(-1*time.Hour)))

	repo := sqlite.NewTaskRepository(db)
	uc := NewNavQueryUseCase(repo)

	tasks, err := uc.ListByView(context.Background(), domain.ViewToday, now, "")
	if err != nil {
		t.Fatalf("list today: %v", err)
	}
	got := titles(tasks)
	assertContains(t, got, "doing-a")
	assertContains(t, got, "todo-overdue")
	assertContains(t, got, "todo-today")
	assertNotContains(t, got, "todo-future")
	assertNotContains(t, got, "inbox-today")
	assertNotContains(t, got, "done-overdue")
}

func openNavTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	return db
}

func seedNavTask(t *testing.T, db *sql.DB, title string, status domain.Status, project string, dueAt, doneAt *time.Time) {
	t.Helper()
	var due any
	if dueAt != nil {
		due = *dueAt
	}
	var done any
	if doneAt != nil {
		done = *doneAt
	}
	_, err := db.Exec(
		`INSERT INTO tasks(title, notes, status, project, priority, due_at, done_at)
		 VALUES (?, '', ?, ?, 'P2', ?, ?)`,
		title, string(status), project, due, done,
	)
	if err != nil {
		t.Fatalf("seed task %q: %v", title, err)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func titles(tasks []domain.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, task.Title)
	}
	return out
}

func assertContains(t *testing.T, list []string, want string) {
	t.Helper()
	for _, item := range list {
		if item == want {
			return
		}
	}
	t.Fatalf("list %v should contain %q", list, want)
}

func assertNotContains(t *testing.T, list []string, target string) {
	t.Helper()
	for _, item := range list {
		if item == target {
			t.Fatalf("list %v should not contain %q", list, target)
		}
	}
}
