package cli

import (
	"fmt"
	"strings"
	"time"
)

func parseDueInput(raw string, loc *time.Location) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, fmt.Errorf("due datetime is empty")
	}
	if loc == nil {
		loc = time.Local
	}
	if t, err := time.Parse(time.RFC3339, text); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("200601021504", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", text, loc); err == nil {
		dayEnd := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, loc)
		return dayEnd.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid due datetime %q, expect YYYY-MM-DD, YYYY-MM-DD HH:MM, or YYYYMMDDHHMM", raw)
}
