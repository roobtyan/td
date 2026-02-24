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
	parsed, _, err := u.ParseInput(ctx, text, useAI)
	if err != nil {
		return domain.Task{}, err
	}
	return u.CreateFromParsed(ctx, parsed)
}

func (u AddFromClipboardUseCase) ParseInput(ctx context.Context, text string, useAI bool) (clipboard.ParsedTask, string, error) {
	if strings.TrimSpace(text) == "" {
		reader := u.ReadClipboard
		if reader == nil {
			reader = clipboard.ReadText
		}
		raw, err := reader()
		if err != nil {
			return clipboard.ParsedTask{}, "", err
		}
		text = raw
	}

	parsed := clipboard.ParseByRule(text)
	source := "fallback"
	if useAI && u.AIParser != nil {
		aiParsed, aiSource, err := u.AIParser.ParseTaskWithSource(ctx, text)
		if err == nil {
			parsed = aiParsed
			source = aiSource
		}
	}
	if strings.TrimSpace(parsed.Title) == "" {
		return clipboard.ParsedTask{}, "", errors.New("clipboard text is empty")
	}
	if strings.TrimSpace(parsed.Priority) == "" {
		parsed.Priority = "P2"
	}
	return parsed, source, nil
}

func (u AddFromClipboardUseCase) CreateFromParsed(ctx context.Context, parsed clipboard.ParsedTask) (domain.Task, error) {
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
	dueAt := u.DueAt
	if dueAt == nil {
		if parsedDue, ok := parseDueText(parsed.Due, time.Local); ok {
			dueAt = &parsedDue
		}
	}
	status := domain.StatusInbox
	if project != "" || dueAt != nil {
		status = domain.StatusTodo
	}
	id, err := u.Repo.Create(ctx, domain.Task{
		Title:    parsed.Title,
		Notes:    parsed.Notes,
		Status:   status,
		Project:  project,
		Priority: priority,
		DueAt:    dueAt,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return u.Repo.GetByID(ctx, id)
}

func parseDueText(raw string, loc *time.Location) (time.Time, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, text); err == nil {
		return t.In(loc), true
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", text, loc); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", text, loc); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("200601021504", text, loc); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", text, loc); err == nil {
		return t, true
	}
	return time.Time{}, false
}
