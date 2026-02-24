package clipboard

import (
	"regexp"
	"strings"
)

var linkRegexp = regexp.MustCompile(`https?://[^\s]+`)

type ParsedTask struct {
	Title    string
	Notes    string
	Project  string
	Priority string
	Due      string
	Links    []string
}

func ExtractLinks(text string) []string {
	matches := linkRegexp.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, link := range matches {
		if _, ok := seen[link]; ok {
			continue
		}
		seen[link] = struct{}{}
		out = append(out, link)
	}
	return out
}

func ParseByRule(raw string) ParsedTask {
	normalized, _ := Normalize(raw)
	links := ExtractLinks(normalized)

	title := firstNonEmptyLine(normalized)
	if title == "" {
		if len(links) > 0 {
			title = links[0]
		} else {
			title = "clipboard task"
		}
	}

	return ParsedTask{
		Title:    title,
		Notes:    normalized,
		Priority: "P2",
		Links:    links,
	}
}

func firstNonEmptyLine(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean != "" {
			return clean
		}
	}
	return ""
}
