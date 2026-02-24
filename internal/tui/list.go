package tui

import (
	"fmt"
	"strings"
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
	label := "[" + string(status) + "]"
	switch status {
	case domain.StatusInbox:
		return statusInboxStyle.Render(label)
	case domain.StatusTodo:
		return statusTodoStyle.Render(label)
	case domain.StatusDoing:
		return statusDoingStyle.Render(label)
	case domain.StatusDone:
		return statusDoneStyle.Render(label)
	case domain.StatusDeleted:
		return statusDelStyle.Render(label)
	default:
		return listRowStyle.Render(label)
	}
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
	meta := renderTaskMeta(task, view, loc)
	return composeTaskLine(prefix, statusField, task.Title, meta, width)
}

func renderTaskMeta(task domain.Task, view domain.View, loc *time.Location) string {
	segments := make([]string, 0, 3)

	switch view {
	case domain.ViewLog:
		segments = append(segments, renderProjectMeta(task.Project))
		segments = append(segments, renderDoneMeta(task.DoneAt, loc))
		segments = append(segments, renderPriorityMeta(task.Priority))
	case domain.ViewTrash:
		segments = append(segments, renderProjectMeta(task.Project))
		segments = append(segments, renderPriorityMeta(task.Priority))
	case domain.ViewToday:
		segments = append(segments, renderProjectMeta(task.Project))
		segments = append(segments, renderDueMeta(task.DueAt, loc))
		segments = append(segments, renderPriorityMeta(task.Priority))
	default:
		segments = append(segments, renderDueMeta(task.DueAt, loc))
		segments = append(segments, renderPriorityMeta(task.Priority))
	}
	return strings.Join(segments, "  ")
}

func renderProjectMeta(project string) string {
	project = strings.TrimSpace(project)
	if project == "" {
		return metaMutedStyle.Render("-")
	}
	return metaProjectStyle.Render(project)
}

func renderDueMeta(dueAt *time.Time, loc *time.Location) string {
	if dueAt == nil {
		return metaMutedStyle.Render("-")
	}
	if loc == nil {
		loc = time.Local
	}
	dueLocal := dueAt.In(loc)
	label := dueLocal.Format("2006-01-02 15:04")
	if dueLocal.Before(time.Now().In(loc)) {
		return metaDueOverdueStyle.Render(label)
	}
	return metaDueStyle.Render(label)
}

func renderDoneMeta(doneAt *time.Time, loc *time.Location) string {
	if doneAt == nil {
		return metaMutedStyle.Render("-")
	}
	return metaDoneStyle.Render(formatDoneAt(doneAt, loc))
}

func renderPriorityMeta(priority string) string {
	priority = domain.NormalizePriority(priority)
	if !domain.IsValidPriority(priority) {
		priority = domain.DefaultPriority
	}
	switch priority {
	case "P1":
		return priorityP1Style.Render(priority)
	case "P2":
		return priorityP2Style.Render(priority)
	case "P3":
		return priorityP3Style.Render(priority)
	default:
		return priorityP4Style.Render(priority)
	}
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
