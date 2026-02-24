package repo

import (
	"context"
	"time"

	"td/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (domain.Task, error)
	List(ctx context.Context, filter TaskListFilter) ([]domain.Task, error)
	CreateProject(ctx context.Context, name string) error
	ListProjects(ctx context.Context) ([]string, error)
	RenameProject(ctx context.Context, oldName, newName string) error
	DeleteProject(ctx context.Context, name string) error
	UpdateTitle(ctx context.Context, id int64, title string) error
	UpdateProject(ctx context.Context, id int64, project string) error
	UpdateDueAt(ctx context.Context, id int64, dueAt *time.Time) error
	UpdatePriority(ctx context.Context, id int64, priority string) error
	SetStatus(ctx context.Context, id int64, status domain.Status) error
	MarkDone(ctx context.Context, ids []int64) error
	MarkDoing(ctx context.Context, ids []int64) error
	Reopen(ctx context.Context, ids []int64) error
	SoftDelete(ctx context.Context, ids []int64) error
	Restore(ctx context.Context, ids []int64) error
	Purge(ctx context.Context, ids []int64) error
}

type TaskListFilter struct {
	Project string
	Limit   int
}
