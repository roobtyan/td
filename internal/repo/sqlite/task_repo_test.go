package sqlite

import (
	"context"
	"testing"

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
