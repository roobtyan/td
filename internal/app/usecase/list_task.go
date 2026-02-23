package usecase

import (
	"context"

	"td/internal/domain"
	"td/internal/repo"
)

type ListTaskInput struct {
	Project string
	Limit   int
}

type ListTaskUseCase struct {
	Repo repo.TaskRepository
}

func (u ListTaskUseCase) Execute(ctx context.Context, in ListTaskInput) ([]domain.Task, error) {
	return u.Repo.List(ctx, repo.TaskListFilter{
		Project: in.Project,
		Limit:   in.Limit,
	})
}
