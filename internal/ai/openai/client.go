package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"td/internal/ai"
)

type Client struct {
	Endpoint   string
	APIKey     string
	Model      string
	Cache      *ai.Cache
	HTTPClient *http.Client
}

func (c *Client) ParseTask(ctx context.Context, input string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("openai api key is empty")
	}
	endpoint := resolveChatCompletionsEndpoint(c.Endpoint)
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "deepseek-chat"
	}
	nowText := time.Now().Local().Format("2006-01-02 15:04")

	cacheKey := model + "\n" + input
	if c.Cache != nil {
		if cached, ok := c.Cache.Get(cacheKey); ok {
			return cached, nil
		}
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "Extract one todo from user text. Return only JSON with keys title,notes,project,priority,due,links. " +
					"priority must be one of P1,P2,P3,P4. due must be empty string or local datetime in YYYY-MM-DD HH:MM. " +
					"Extract project name when text indicates ownership, such as 在XXX项目下/归属XXX/for XXX project; otherwise project should be empty. " +
					"Resolve relative time phrases (today/tomorrow/明天) using local time " + nowText + ".",
			},
			{
				"role":    "user",
				"content": input,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("openai api error: %s", extractAPIError(respBody, resp.Status))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", errors.New("openai api returned empty choices")
	}
	content := trimCodeFence(result.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("openai api returned empty content")
	}
	if c.Cache != nil {
		c.Cache.Put(cacheKey, content)
	}
	return content, nil
}

func resolveChatCompletionsEndpoint(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "https://api.deepseek.com/v1/chat/completions"
	}
	text = strings.TrimRight(text, "/")
	if strings.HasSuffix(text, "/chat/completions") {
		return text
	}
	if strings.HasSuffix(text, "/v1") {
		return text + "/chat/completions"
	}
	return text + "/chat/completions"
}

func trimCodeFence(raw string) string {
	text := strings.TrimSpace(raw)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return text
	}
	lines = lines[1:]
	last := strings.TrimSpace(lines[len(lines)-1])
	if strings.HasPrefix(last, "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func extractAPIError(body []byte, fallback string) string {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		if msg := strings.TrimSpace(payload.Error.Message); msg != "" {
			return msg
		}
		if msg := strings.TrimSpace(payload.Message); msg != "" {
			return msg
		}
	}
	raw := strings.TrimSpace(string(body))
	if raw != "" {
		return raw
	}
	return fallback
}
