package usecase

import (
	"context"

	"td/internal/domain"
	"td/internal/repo"
)

type AddTaskInput struct {
	Title    string
	Project  string
	Priority string
}

type AddTaskUseCase struct {
	Repo repo.TaskRepository
}

func (u AddTaskUseCase) Execute(ctx context.Context, in AddTaskInput) (domain.Task, error) {
	priority := in.Priority
	if priority == "" {
		priority = "P2"
	}

	id, err := u.Repo.Create(ctx, domain.Task{
		Title:    in.Title,
		Status:   domain.StatusInbox,
		Project:  in.Project,
		Priority: priority,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return u.Repo.GetByID(ctx, id)
}
