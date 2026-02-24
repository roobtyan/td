package sqlite

import (
	"context"
	"testing"
	"time"

	"td/internal/domain"
)

func TestTaskLifecycle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "write plan",
		Status: domain.StatusTodo,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.MarkDone(ctx, []int64{id}); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != domain.StatusDone {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusDone)
	}

	if err := repo.Reopen(ctx, []int64{id}); err != nil {
		t.Fatalf("reopen task: %v", err)
	}
	task, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after reopen: %v", err)
	}
	if task.Status != domain.StatusTodo {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusTodo)
	}
}

func TestTaskMarkDoing(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "start task",
		Status: domain.StatusInbox,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.MarkDoing(ctx, []int64{id}); err != nil {
		t.Fatalf("mark doing: %v", err)
	}
	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after mark doing: %v", err)
	}
	if task.Status != domain.StatusDoing {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusDoing)
	}
}

func TestTaskUpdateProjectAndDueAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "plan sprint",
		Status: domain.StatusTodo,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.UpdateProject(ctx, id, "work"); err != nil {
		t.Fatalf("update project: %v", err)
	}
	due := time.Date(2026, 2, 24, 20, 30, 0, 0, time.FixedZone("CST", 8*3600))
	if err := repo.UpdateDueAt(ctx, id, &due); err != nil {
		t.Fatalf("update due: %v", err)
	}

	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after updates: %v", err)
	}
	if task.Project != "work" {
		t.Fatalf("project = %q, want %q", task.Project, "work")
	}
	if task.DueAt == nil {
		t.Fatalf("due_at should not be nil")
	}
	if !task.DueAt.Equal(due.UTC()) {
		t.Fatalf("due_at = %s, want %s", task.DueAt.UTC(), due.UTC())
	}
}

func TestProjectCRUDAndTaskDetachOnDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	if err := repo.CreateProject(ctx, "work"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := repo.CreateProject(ctx, "home"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	projects, err := repo.ListProjects(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 2 || projects[0] != "home" || projects[1] != "work" {
		t.Fatalf("projects = %v, want [home work]", projects)
	}

	id, err := repo.Create(ctx, domain.Task{
		Title:   "task in work",
		Status:  domain.StatusTodo,
		Project: "work",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.RenameProject(ctx, "work", "office"); err != nil {
		t.Fatalf("rename project: %v", err)
	}
	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Project != "office" {
		t.Fatalf("task project = %q, want %q", task.Project, "office")
	}

	if err := repo.DeleteProject(ctx, "office"); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	task, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after delete project: %v", err)
	}
	if task.Project != "" {
		t.Fatalf("task project = %q, want empty", task.Project)
	}
}

func TestUpdateProjectShouldMoveInboxToTodo(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "inbox task",
		Status: domain.StatusInbox,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := repo.UpdateProject(ctx, id, "work"); err != nil {
		t.Fatalf("update project: %v", err)
	}

	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Project != "work" {
		t.Fatalf("project = %q, want %q", task.Project, "work")
	}
	if task.Status != domain.StatusTodo {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusTodo)
	}
}

func TestRestoreShouldMoveTaskWithoutProjectToInbox(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "trash task",
		Status: domain.StatusInbox,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := repo.SoftDelete(ctx, []int64{id}); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := repo.Restore(ctx, []int64{id}); err != nil {
		t.Fatalf("restore: %v", err)
	}

	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != domain.StatusInbox {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusInbox)
	}
}

func TestSetStatusShouldUpdateStatusAndDoneAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewTaskRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.Task{
		Title:  "status task",
		Status: domain.StatusInbox,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.SetStatus(ctx, id, domain.StatusDone); err != nil {
		t.Fatalf("set status done: %v", err)
	}
	task, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after set done: %v", err)
	}
	if task.Status != domain.StatusDone {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusDone)
	}
	if task.DoneAt == nil {
		t.Fatalf("done_at should not be nil when status is done")
	}

	if err := repo.SetStatus(ctx, id, domain.StatusDoing); err != nil {
		t.Fatalf("set status doing: %v", err)
	}
	task, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get task after set doing: %v", err)
	}
	if task.Status != domain.StatusDoing {
		t.Fatalf("status = %s, want %s", task.Status, domain.StatusDoing)
	}
	if task.DoneAt != nil {
		t.Fatalf("done_at should be nil when status is not done")
	}
}
