package repo

import (
	"context"

	"td/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (domain.Task, error)
	List(ctx context.Context, filter TaskListFilter) ([]domain.Task, error)
	MarkDone(ctx context.Context, ids []int64) error
	Reopen(ctx context.Context, ids []int64) error
}

type TaskListFilter struct {
	Project string
	Limit   int
}
