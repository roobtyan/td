package usecase

import (
	"context"
	"strings"

	"td/internal/ai"
	"td/internal/ai/schema"
	"td/internal/clipboard"
)

type AIParseTaskUseCase struct {
	Provider ai.Provider
}

func (u AIParseTaskUseCase) ParseTask(ctx context.Context, input string) (clipboard.ParsedTask, error) {
	parsed, _, err := u.ParseTaskWithSource(ctx, input)
	return parsed, err
}

func (u AIParseTaskUseCase) ParseTaskWithSource(ctx context.Context, input string) (clipboard.ParsedTask, string, error) {
	fallback := clipboard.ParseByRule(input)
	if u.Provider == nil {
		return fallback, "fallback", nil
	}

	sanitized := ai.RedactInput(input)
	raw, err := u.Provider.ParseTask(ctx, sanitized)
	if err != nil {
		return fallback, "fallback", nil
	}
	payload, err := schema.DecodeParseTaskJSON(raw)
	if err != nil {
		return fallback, "fallback", nil
	}

	parsed := clipboard.ParsedTask{
		Title:    strings.TrimSpace(payload.Title),
		Notes:    strings.TrimSpace(payload.Notes),
		Project:  strings.TrimSpace(payload.Project),
		Priority: strings.TrimSpace(payload.Priority),
		Due:      strings.TrimSpace(payload.Due),
		Links:    payload.Links,
	}
	if parsed.Title == "" {
		return fallback, "fallback", nil
	}
	if parsed.Notes == "" {
		parsed.Notes = fallback.Notes
	}
	if parsed.Priority == "" {
		parsed.Priority = "P2"
	}
	if len(parsed.Links) == 0 {
		parsed.Links = fallback.Links
	}
	return parsed, "ai", nil
}
