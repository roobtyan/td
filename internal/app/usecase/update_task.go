package usecase

import (
	"context"

	"td/internal/repo"
)

type UpdateTaskUseCase struct {
	Repo repo.TaskRepository
}

func (u UpdateTaskUseCase) MarkDone(ctx context.Context, ids []int64) error {
	return u.Repo.MarkDone(ctx, ids)
}

func (u UpdateTaskUseCase) Reopen(ctx context.Context, ids []int64) error {
	return u.Repo.Reopen(ctx, ids)
}

func (u UpdateTaskUseCase) EditTitle(ctx context.Context, id int64, title string) error {
	return u.Repo.UpdateTitle(ctx, id, title)
}

func (u UpdateTaskUseCase) Remove(ctx context.Context, ids []int64) error {
	return u.Repo.SoftDelete(ctx, ids)
}

func (u UpdateTaskUseCase) Restore(ctx context.Context, ids []int64) error {
	return u.Repo.Restore(ctx, ids)
}

func (u UpdateTaskUseCase) Purge(ctx context.Context, ids []int64) error {
	return u.Repo.Purge(ctx, ids)
}
