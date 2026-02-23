package clipboard

import "strings"

const MaxClipboardChars = 4000

func Normalize(raw string) (string, bool) {
	text := strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	out := make([]string, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			if prevBlank {
				continue
			}
			prevBlank = true
			out = append(out, "")
			continue
		}
		prevBlank = false
		out = append(out, clean)
	}

	normalized := strings.TrimSpace(strings.Join(out, "\n"))
	if len(normalized) <= MaxClipboardChars {
		return normalized, false
	}
	return normalized[:MaxClipboardChars], true
}
