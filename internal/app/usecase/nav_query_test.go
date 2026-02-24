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

	tasks, err := uc.ListByView(context.Background(), domain.ViewToday, now, "", false)
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

func TestProjectViewShowDoneToggle(t *testing.T) {
	db := openNavTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	seedNavTask(t, db, "work-inbox", domain.StatusInbox, "work", nil, nil)
	seedNavTask(t, db, "work-doing", domain.StatusDoing, "work", nil, nil)
	seedNavTask(t, db, "work-done", domain.StatusDone, "work", nil, ptrTime(now))
	seedNavTask(t, db, "home-todo", domain.StatusTodo, "home", nil, nil)

	repo := sqlite.NewTaskRepository(db)
	uc := NewNavQueryUseCase(repo)

	hiddenDone, err := uc.ListByView(context.Background(), domain.ViewProject, now, "work", false)
	if err != nil {
		t.Fatalf("list project hidden done: %v", err)
	}
	gotHidden := titles(hiddenDone)
	assertContains(t, gotHidden, "work-inbox")
	assertContains(t, gotHidden, "work-doing")
	assertNotContains(t, gotHidden, "work-done")

	withDone, err := uc.ListByView(context.Background(), domain.ViewProject, now, "work", true)
	if err != nil {
		t.Fatalf("list project with done: %v", err)
	}
	gotAll := titles(withDone)
	assertContains(t, gotAll, "work-inbox")
	assertContains(t, gotAll, "work-doing")
	assertContains(t, gotAll, "work-done")
	assertNotContains(t, gotAll, "home-todo")
}

func TestInboxViewShouldIncludeTodoWithoutProject(t *testing.T) {
	db := openNavTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	seedNavTask(t, db, "inbox-a", domain.StatusInbox, "", nil, nil)
	seedNavTask(t, db, "todo-no-project", domain.StatusTodo, "", nil, nil)
	seedNavTask(t, db, "todo-with-project", domain.StatusTodo, "work", nil, nil)
	seedNavTask(t, db, "doing-no-project", domain.StatusDoing, "", nil, nil)

	repo := sqlite.NewTaskRepository(db)
	uc := NewNavQueryUseCase(repo)

	tasks, err := uc.ListByView(context.Background(), domain.ViewInbox, now, "", false)
	if err != nil {
		t.Fatalf("list inbox: %v", err)
	}
	got := titles(tasks)
	assertContains(t, got, "inbox-a")
	assertContains(t, got, "todo-no-project")
	assertNotContains(t, got, "todo-with-project")
	assertNotContains(t, got, "doing-no-project")
}

func TestLogViewShouldSortByLatestDoneAt(t *testing.T) {
	db := openNavTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	seedNavTask(t, db, "done-old", domain.StatusDone, "work", nil, ptrTime(now.Add(-3*time.Hour)))
	seedNavTask(t, db, "done-latest", domain.StatusDone, "work", nil, ptrTime(now.Add(-30*time.Minute)))
	seedNavTask(t, db, "done-middle", domain.StatusDone, "work", nil, ptrTime(now.Add(-90*time.Minute)))

	repo := sqlite.NewTaskRepository(db)
	uc := NewNavQueryUseCase(repo)

	tasks, err := uc.ListByView(context.Background(), domain.ViewLog, now, "", false)
	if err != nil {
		t.Fatalf("list log: %v", err)
	}
	got := titles(tasks)
	want := []string{"done-latest", "done-middle", "done-old"}
	if len(got) != len(want) {
		t.Fatalf("log size = %d, want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("log order mismatch at %d, got=%v want=%v", i, got, want)
		}
	}
}

func TestTrashViewShouldSortByLatestDeletedAt(t *testing.T) {
	db := openNavTestDB(t)
	defer db.Close()
	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	seedNavTask(t, db, "trash-old", domain.StatusDeleted, "work", nil, nil)
	seedNavTask(t, db, "trash-latest", domain.StatusDeleted, "work", nil, nil)
	seedNavTask(t, db, "trash-middle", domain.StatusDeleted, "work", nil, nil)
	seedNavTask(t, db, "not-trash", domain.StatusTodo, "work", nil, nil)

	setNavTaskUpdatedAtByTitle(t, db, "trash-old", now.Add(-3*time.Hour))
	setNavTaskUpdatedAtByTitle(t, db, "trash-latest", now.Add(-30*time.Minute))
	setNavTaskUpdatedAtByTitle(t, db, "trash-middle", now.Add(-90*time.Minute))

	repo := sqlite.NewTaskRepository(db)
	uc := NewNavQueryUseCase(repo)

	tasks, err := uc.ListByView(context.Background(), domain.ViewTrash, now, "", false)
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	got := titles(tasks)
	want := []string{"trash-latest", "trash-middle", "trash-old"}
	if len(got) != len(want) {
		t.Fatalf("trash size = %d, want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("trash order mismatch at %d, got=%v want=%v", i, got, want)
		}
	}
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

func setNavTaskUpdatedAtByTitle(t *testing.T, db *sql.DB, title string, updatedAt time.Time) {
	t.Helper()
	_, err := db.Exec(`UPDATE tasks SET updated_at = ? WHERE title = ?`, updatedAt, title)
	if err != nil {
		t.Fatalf("set updated_at for %q: %v", title, err)
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
