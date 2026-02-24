package usecase

import (
	"context"
	"time"

	"td/internal/domain"
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

func (u UpdateTaskUseCase) SetProject(ctx context.Context, id int64, project string) error {
	return u.Repo.UpdateProject(ctx, id, project)
}

func (u UpdateTaskUseCase) SetDueAt(ctx context.Context, id int64, dueAt *time.Time) error {
	return u.Repo.UpdateDueAt(ctx, id, dueAt)
}

func (u UpdateTaskUseCase) SetPriority(ctx context.Context, id int64, priority string) error {
	return u.Repo.UpdatePriority(ctx, id, priority)
}

func (u UpdateTaskUseCase) SetStatus(ctx context.Context, id int64, status domain.Status) error {
	return u.Repo.SetStatus(ctx, id, status)
}

func (u UpdateTaskUseCase) MarkToday(ctx context.Context, ids []int64) error {
	return u.Repo.MarkDoing(ctx, ids)
}

func (u UpdateTaskUseCase) MarkProjectDone(ctx context.Context, project string) (int, error) {
	tasks, err := u.Repo.List(ctx, repo.TaskListFilter{Project: project})
	if err != nil {
		return 0, err
	}
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == domain.StatusInbox || task.Status == domain.StatusTodo || task.Status == domain.StatusDoing {
			ids = append(ids, task.ID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := u.Repo.MarkDone(ctx, ids); err != nil {
		return 0, err
	}
	return len(ids), nil
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
