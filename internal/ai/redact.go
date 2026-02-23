package ai

import (
	"regexp"
	"strings"
)

var redactURLRegexp = regexp.MustCompile(`https?://[^\s]+`)

func RedactInput(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	return redactURLRegexp.ReplaceAllString(trimmed, "[URL]")
}
