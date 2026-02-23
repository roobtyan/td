package usecase

import (
	"context"
	"time"

	"td/internal/domain"
	"td/internal/repo"
)

type AddTaskInput struct {
	Title    string
	Project  string
	Priority string
	DueAt    *time.Time
}

type AddTaskUseCase struct {
	Repo repo.TaskRepository
}

func (u AddTaskUseCase) Execute(ctx context.Context, in AddTaskInput) (domain.Task, error) {
	priority := in.Priority
	if priority == "" {
		priority = "P2"
	}
	status := domain.StatusInbox
	if in.Project != "" {
		status = domain.StatusTodo
	}

	id, err := u.Repo.Create(ctx, domain.Task{
		Title:    in.Title,
		Status:   status,
		Project:  in.Project,
		Priority: priority,
		DueAt:    in.DueAt,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return u.Repo.GetByID(ctx, id)
}
