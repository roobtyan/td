package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/x/ansi"

	"td/internal/domain"
)

func renderList(tasks []domain.Task, cursor int, focused bool, width, height int, view domain.View, loc *time.Location) []string {
	contentWidth := paneContentWidth(width)
	contentHeight := paneContentHeight(height)
	lines := []string{truncateLineForPane("Tasks", contentWidth)}
	if contentHeight > 1 {
		if len(tasks) == 0 {
			lines = append(lines, truncateLineForPane("[empty] no tasks in this view", contentWidth))
		} else {
			visible := contentHeight - 1
			start, end := viewportWindow(len(tasks), cursor, visible)
			for idx := start; idx < end; idx++ {
				task := tasks[idx]
				prefix := "  "
				if idx == cursor {
					prefix = "> "
				}
				status := renderStatusLabel(task.Status)
				line := renderTaskLine(prefix, status, task, view, loc, contentWidth)
				if idx == cursor {
					line = listSelectedStyle.Render(line)
				}
				lines = append(lines, line)
			}
		}
	}
	style := listBoxStyle.Copy()
	if focused {
		style = style.BorderForeground(focusColor)
	}
	return splitRendered(renderBox(style, joinLines(lines), width, height))
}

func renderStatusLabel(status domain.Status) string {
	return "[" + string(status) + "]"
}

func formatDue(dueAt *time.Time, loc *time.Location) string {
	if dueAt == nil {
		return "-"
	}
	if loc == nil {
		loc = time.Local
	}
	return dueAt.In(loc).Format("2006-01-02 15:04")
}

func formatDoneAt(doneAt *time.Time, loc *time.Location) string {
	if doneAt == nil {
		return "-"
	}
	if loc == nil {
		loc = time.Local
	}
	return doneAt.In(loc).Format("2006-01-02 15:04")
}

func renderTaskLine(prefix, status string, task domain.Task, view domain.View, loc *time.Location, width int) string {
	if width <= 0 {
		width = 1
	}
	if loc == nil {
		loc = time.Local
	}
	statusField := padFixed(status, 9)

	if view == domain.ViewLog {
		project := task.Project
		if project == "" {
			project = "-"
		}
		meta := fmt.Sprintf("proj:%s  done:%s", project, formatDoneAt(task.DoneAt, loc))
		return composeTaskLine(prefix, statusField, task.Title, meta, width)
	}
	if view == domain.ViewTrash {
		project := task.Project
		if project == "" {
			project = "-"
		}
		meta := fmt.Sprintf("proj:%s", project)
		return composeTaskLine(prefix, statusField, task.Title, meta, width)
	}
	if view == domain.ViewToday {
		project := task.Project
		if project == "" {
			project = "-"
		}
		meta := fmt.Sprintf("proj:%s  due:%s", project, formatDue(task.DueAt, loc))
		return composeTaskLine(prefix, statusField, task.Title, meta, width)
	}
	meta := fmt.Sprintf("due:%s", formatDue(task.DueAt, loc))
	return composeTaskLine(prefix, statusField, task.Title, meta, width)
}

func composeTaskLine(prefix, statusField, title, meta string, width int) string {
	base := prefix + " " + statusField + " "
	baseWidth := ansi.StringWidth(base)
	metaWidth := ansi.StringWidth(meta)
	availableTitle := width - baseWidth - 2 - metaWidth
	if availableTitle < 4 {
		availableTitle = 4
	}
	title = ansi.Truncate(flattenPaneLine(title), availableTitle, "...")
	title = padFixed(title, availableTitle)
	return truncateLineForPane(base+title+"  "+meta, width)
}

func padFixed(text string, width int) string {
	w := ansi.StringWidth(text)
	if w >= width {
		return ansi.Truncate(text, width, "")
	}
	return text + spaces(width-w)
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}
