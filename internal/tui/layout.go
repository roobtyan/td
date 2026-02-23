package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"td/internal/domain"
)

func bodyPaneWidths(totalWidth int) (int, int) {
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

func joinSections(header, body, footer string) string {
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
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

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func splitRendered(block string) []string {
	return strings.Split(block, "\n")
}

func renderHeader(done, total int, view domain.View, count int, now time.Time) string {
	progressBar := renderProgressBar(done, total, 18)
	percent := progressPercent(done, total)
	progressLine := fmt.Sprintf(
		"Today Progress %s %d/%d (%d%%)  View: %s  Items: %d  %s",
		progressBar,
		done,
		total,
		percent,
		viewLabel(view),
		count,
		now.Format("15:04"),
	)

	logoLines := []string{
		" _____ _____ ",
		"|_   _|  _  |",
		"  | | | | | |",
		"  |_| |_| |_|",
		"      TD     ",
	}
	logo := logoStyle.Render(strings.Join(logoLines, "\n"))
	info := headerInfoStyle.Render(progressLine)
	return headerBoxStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, logo, "  ", info),
	)
}

func renderFooter(statusMsg string) string {
	status := statusMsg
	if status == "" {
		status = "ready"
	}
	left := footerStatusStyle.Render(status)
	right := footerHintStyle.Render("j/k move  Tab focus  Enter select  p clip  Ctrl+a ai  q quit")
	return footerBoxStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right))
}

func progressPercent(done, total int) int {
	if total <= 0 || done <= 0 {
		return 0
	}
	return int(float64(done) / float64(total) * 100)
}

func renderProgressBar(done, total, width int) string {
	if width <= 0 {
		width = 12
	}
	fill := 0
	if total > 0 && done > 0 {
		fill = int(float64(done) / float64(total) * float64(width))
	}
	if fill > width {
		fill = width
	}
	if fill < 0 {
		fill = 0
	}
	return "[" + strings.Repeat("#", fill) + strings.Repeat("-", width-fill) + "]"
}

func viewLabel(v domain.View) string {
	switch v {
	case domain.ViewToday:
		return "Today"
	case domain.ViewInbox:
		return "Inbox"
	case domain.ViewLog:
		return "Log"
	case domain.ViewProject:
		return "Project"
	case domain.ViewTrash:
		return "Trash"
	default:
		return string(v)
	}
}
