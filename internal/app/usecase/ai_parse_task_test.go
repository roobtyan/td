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

type fakeParseProvider struct {
	raw string
	err error
}

func (f fakeParseProvider) ParseTask(_ context.Context, _ string) (string, error) {
	return f.raw, f.err
}
