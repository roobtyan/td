package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"td/internal/clipboard"
	"td/internal/domain"
	"td/internal/repo"
)

type AddFromClipboardUseCase struct {
	Repo          repo.TaskRepository
	ReadClipboard func() (string, error)
	AIParser      *AIParseTaskUseCase
	Project       string
	Priority      string
	DueAt         *time.Time
}

func (u AddFromClipboardUseCase) AddFromClipboard(ctx context.Context, text string, useAI bool) (domain.Task, error) {
	if strings.TrimSpace(text) == "" {
		reader := u.ReadClipboard
		if reader == nil {
			reader = clipboard.ReadText
		}
		raw, err := reader()
		if err != nil {
			return domain.Task{}, err
		}
		text = raw
	}

	parsed := clipboard.ParseByRule(text)
	if useAI && u.AIParser != nil {
		aiParsed, err := u.AIParser.ParseTask(ctx, text)
		if err == nil {
			parsed = aiParsed
		}
	}
	if strings.TrimSpace(parsed.Title) == "" {
		return domain.Task{}, errors.New("clipboard text is empty")
	}

	priority := u.Priority
	if priority == "" {
		priority = parsed.Priority
		if priority == "" {
			priority = "P2"
		}
	}
	project := u.Project
	if project == "" {
		project = parsed.Project
	}
	status := domain.StatusInbox
	if project != "" {
		status = domain.StatusTodo
	}
	id, err := u.Repo.Create(ctx, domain.Task{
		Title:    parsed.Title,
		Notes:    parsed.Notes,
		Status:   status,
		Project:  project,
		Priority: priority,
		DueAt:    u.DueAt,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return u.Repo.GetByID(ctx, id)
}
