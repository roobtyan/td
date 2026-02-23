package tui

import (
	"strings"
)

func paneWidths(totalWidth int) (int, int) {
	if totalWidth <= 0 {
		totalWidth = 80
	}
	navWidth := totalWidth / 4
	if navWidth < 20 {
		navWidth = 20
	}
	listWidth := totalWidth - navWidth - 3
	if listWidth < 20 {
		listWidth = 20
		navWidth = totalWidth - listWidth - 3
		if navWidth < 10 {
			navWidth = 10
		}
	}
	return navWidth, listWidth
}

func joinColumns(left, right []string, leftWidth, rightWidth int) string {
	maxLines := len(left)
	if len(right) > maxLines {
		maxLines = len(right)
	}
	var b strings.Builder
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}
		b.WriteString(padRight(leftLine, leftWidth))
		b.WriteString(" | ")
		b.WriteString(padRight(rightLine, rightWidth))
		if i != maxLines-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func padRight(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(text) > width {
		if width <= 3 {
			return text[:width]
		}
		return text[:width-3] + "..."
	}
	if len(text) == width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}
