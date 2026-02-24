package usecase

import (
	"context"
	"sort"
	"strings"
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

func (u NavQueryUseCase) ListByView(ctx context.Context, view domain.View, now time.Time, project string, includeDone bool) ([]domain.Task, error) {
	tasks, err := u.Repo.List(ctx, repo.TaskListFilter{})
	if err != nil {
		return nil, err
	}

	out := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		if u.matchView(task, view, now, project, includeDone) {
			out = append(out, task)
		}
	}
	if view == domain.ViewLog {
		sort.SliceStable(out, func(i, j int) bool {
			left := out[i]
			right := out[j]
			if left.DoneAt == nil && right.DoneAt == nil {
				return left.ID > right.ID
			}
			if left.DoneAt == nil {
				return false
			}
			if right.DoneAt == nil {
				return true
			}
			if left.DoneAt.Equal(*right.DoneAt) {
				return left.ID > right.ID
			}
			return left.DoneAt.After(*right.DoneAt)
		})
	}
	if view == domain.ViewTrash {
		sort.SliceStable(out, func(i, j int) bool {
			left := out[i]
			right := out[j]
			if left.UpdatedAt.Equal(right.UpdatedAt) {
				return left.ID > right.ID
			}
			return left.UpdatedAt.After(right.UpdatedAt)
		})
	}
	return out, nil
}

func (u NavQueryUseCase) matchView(task domain.Task, view domain.View, now time.Time, project string, includeDone bool) bool {
	switch view {
	case domain.ViewInbox:
		if task.Status == domain.StatusInbox {
			return true
		}
		return task.Status == domain.StatusTodo && strings.TrimSpace(task.Project) == ""
	case domain.ViewToday:
		return isTodayTask(task, now)
	case domain.ViewLog:
		return isLogTask(task, now, u.logWindowDays())
	case domain.ViewProject:
		if project == "" {
			return false
		}
		return task.Project == project && isProjectStatus(task.Status, includeDone)
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

func isProjectStatus(status domain.Status, includeDone bool) bool {
	if status == domain.StatusInbox || status == domain.StatusTodo || status == domain.StatusDoing {
		return true
	}
	return includeDone && status == domain.StatusDone
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
