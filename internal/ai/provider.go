package ai

import "context"

type Provider interface {
	ParseTask(ctx context.Context, input string) (string, error)
}
