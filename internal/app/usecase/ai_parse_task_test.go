package usecase

import (
	"context"
	"testing"
)

func TestParseTaskFallback(t *testing.T) {
	uc := AIParseTaskUseCase{
		Provider: fakeParseProvider{
			err: nil,
			raw: `{"title":123}`,
		},
	}

	got, err := uc.ParseTask(context.Background(), "Buy milk\nhttps://example.com")
	if err != nil {
		t.Fatalf("parse task: %v", err)
	}
	if got.Title != "Buy milk" {
		t.Fatalf("fallback title = %q, want %q", got.Title, "Buy milk")
	}
}

func TestParseTaskWithSource(t *testing.T) {
	uc := AIParseTaskUseCase{
		Provider: fakeParseProvider{
			raw: `{"title":"AI title","project":"AI Project","priority":"P2"}`,
		},
	}

	got, source, err := uc.ParseTaskWithSource(context.Background(), "Buy milk")
	if err != nil {
		t.Fatalf("parse task with source: %v", err)
	}
	if source != "ai" {
		t.Fatalf("source = %q, want %q", source, "ai")
	}
	if got.Title != "AI title" {
		t.Fatalf("title = %q, want %q", got.Title, "AI title")
	}
	if got.Project != "AI Project" {
		t.Fatalf("project = %q, want %q", got.Project, "AI Project")
	}
}

type fakeParseProvider struct {
	raw string
	err error
}

func (f fakeParseProvider) ParseTask(_ context.Context, _ string) (string, error) {
	return f.raw, f.err
}
