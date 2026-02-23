package openai

import (
	"context"
	"errors"

	"td/internal/ai"
)

type Client struct {
	Endpoint string
	APIKey   string
	Model    string
	Cache    *ai.Cache
}

func (c *Client) ParseTask(_ context.Context, _ string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("openai api key is empty")
	}
	return "", errors.New("openai parse not implemented")
}
