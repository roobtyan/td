package schema

import (
	"encoding/json"
	"errors"
	"strings"
)

type ParseTaskPayload struct {
	Title    string   `json:"title"`
	Notes    string   `json:"notes"`
	Project  string   `json:"project"`
	Priority string   `json:"priority"`
	Links    []string `json:"links"`
}

func ValidateParseTaskJSON(raw string) error {
	_, err := DecodeParseTaskJSON(raw)
	return err
}

func DecodeParseTaskJSON(raw string) (ParseTaskPayload, error) {
	var payload ParseTaskPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ParseTaskPayload{}, err
	}
	if strings.TrimSpace(payload.Title) == "" {
		return ParseTaskPayload{}, errors.New("title is required")
	}
	if payload.Priority != "" && !isPriority(payload.Priority) {
		return ParseTaskPayload{}, errors.New("invalid priority")
	}
	return payload, nil
}

func isPriority(priority string) bool {
	switch priority {
	case "P1", "P2", "P3", "P4":
		return true
	default:
		return false
	}
}
