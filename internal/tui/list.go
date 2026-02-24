package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"td/internal/domain"
)

const (
	listBaseFG       = "\x1b[38;2;226;232;240m"
	listCursorFG     = "\x1b[1;38;2;63;184;179m"
	listStatusInbox  = "\x1b[38;2;203;213;225m"
	listStatusTodo   = "\x1b[38;2;96;165;250m"
	listStatusDoing  = "\x1b[38;2;251;191;36m"
	listStatusDone   = "\x1b[38;2;52;211;153m"
	listStatusDelete = "\x1b[38;2;248;113;113m"
	listMetaProject  = "\x1b[38;2;150;185;216m"
	listMetaDue      = "\x1b[38;2;125;211;252m"
	listMetaDueWarn  = "\x1b[1;38;2;251;191;36m"
	listMetaDone     = "\x1b[38;2;147;197;253m"
	listMetaMuted    = "\x1b[38;2;147;161;176m"
	listPriP1        = "\x1b[1;38;2;251;113;133m"
	listPriP2        = "\x1b[1;38;2;251;191;36m"
	listPriP3        = "\x1b[38;2;96;165;250m"
	listPriP4        = listMetaMuted
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
				prefix := renderListPrefix(idx == cursor)
				status := renderStatusLabel(task.Status)
				line := renderTaskLine(prefix, status, task, view, loc, contentWidth)
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
		return paintList(label, listStatusInbox)
	case domain.StatusTodo:
		return paintList(label, listStatusTodo)
	case domain.StatusDoing:
		return paintList(label, listStatusDoing)
	case domain.StatusDone:
		return paintList(label, listStatusDone)
	case domain.StatusDeleted:
		return paintList(label, listStatusDelete)
	default:
		return label
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
		return paintList("-", listMetaMuted)
	}
	return paintList(project, listMetaProject)
}

func renderDueMeta(dueAt *time.Time, loc *time.Location) string {
	if dueAt == nil {
		return paintList("-", listMetaMuted)
	}
	if loc == nil {
		loc = time.Local
	}
	dueLocal := dueAt.In(loc)
	label := dueLocal.Format("2006-01-02 15:04")
	if dueLocal.Before(time.Now().In(loc)) {
		return paintList(label, listMetaDueWarn)
	}
	return paintList(label, listMetaDue)
}

func renderDoneMeta(doneAt *time.Time, loc *time.Location) string {
	if doneAt == nil {
		return paintList("-", listMetaMuted)
	}
	return paintList(formatDoneAt(doneAt, loc), listMetaDone)
}

func renderPriorityMeta(priority string) string {
	priority = domain.NormalizePriority(priority)
	if !domain.IsValidPriority(priority) {
		priority = domain.DefaultPriority
	}
	switch priority {
	case "P1":
		return paintList(priority, listPriP1)
	case "P2":
		return paintList(priority, listPriP2)
	case "P3":
		return paintList(priority, listPriP3)
	default:
		return paintList(priority, listPriP4)
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

func renderListPrefix(selected bool) string {
	if selected {
		return paintList("> ", listCursorFG)
	}
	return "  "
}

func paintList(text, colorCode string) string {
	if colorCode == "" {
		return text
	}
	return colorCode + text + listBaseFG
}
