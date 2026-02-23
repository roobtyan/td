package usecase

import (
	"context"
	"time"

	"td/internal/domain"
	"td/internal/repo"
)

type NavQueryUseCase struct {
	Repo          repo.TaskRepository
	LogWindowDays int
}

func NewNavQueryUseCase(repo repo.TaskRepository) NavQueryUseCase {
	return NavQueryUseCase{
		Repo:          repo,
		LogWindowDays: domain.DefaultLogWindowDays,
	}
}

func (u NavQueryUseCase) ListByView(ctx context.Context, view domain.View, now time.Time, project string) ([]domain.Task, error) {
	tasks, err := u.Repo.List(ctx, repo.TaskListFilter{})
	if err != nil {
		return nil, err
	}

	out := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if u.matchView(task, view, now, project) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (u NavQueryUseCase) matchView(task domain.Task, view domain.View, now time.Time, project string) bool {
	switch view {
	case domain.ViewInbox:
		return task.Status == domain.StatusInbox
	case domain.ViewToday:
		return isTodayTask(task, now)
	case domain.ViewLog:
		return isLogTask(task, now, u.logWindowDays())
	case domain.ViewProject:
		if project == "" {
			return false
		}
		return task.Project == project && isProjectStatus(task.Status)
	case domain.ViewTrash:
		return task.Status == domain.StatusDeleted
	default:
		return false
	}
}

func (u NavQueryUseCase) logWindowDays() int {
	if u.LogWindowDays > 0 {
		return u.LogWindowDays
	}
	return domain.DefaultLogWindowDays
}

func isTodayTask(task domain.Task, now time.Time) bool {
	if task.Status != domain.StatusTodo && task.Status != domain.StatusDoing {
		return false
	}
	if task.Status == domain.StatusDoing {
		return true
	}
	if task.DueAt == nil {
		return false
	}

	due := task.DueAt.UTC()
	dayStart := startOfDay(now.UTC())
	dayEnd := dayStart.Add(24 * time.Hour)
	if due.Before(dayStart) {
		return true
	}
	return !due.Before(dayStart) && due.Before(dayEnd)
}

func isLogTask(task domain.Task, now time.Time, windowDays int) bool {
	if task.Status != domain.StatusDone || task.DoneAt == nil {
		return false
	}
	windowStart := now.UTC().Add(-time.Duration(windowDays) * 24 * time.Hour)
	return !task.DoneAt.UTC().Before(windowStart)
}

func isProjectStatus(status domain.Status) bool {
	return status == domain.StatusInbox || status == domain.StatusTodo || status == domain.StatusDoing
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
