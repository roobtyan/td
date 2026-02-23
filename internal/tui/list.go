package tui

import (
	"fmt"

	"td/internal/domain"
)

func renderList(tasks []domain.Task, cursor int, focused bool) []string {
	lines := []string{listTitleStyle.Render("Tasks")}
	if len(tasks) == 0 {
		lines = append(lines, listRowStyle.Render("[empty] no tasks in this view"))
	} else {
		for idx, task := range tasks {
			prefix := "  "
			if idx == cursor {
				prefix = "> "
			}
			status := renderStatusLabel(task.Status)
			line := fmt.Sprintf("%s %s %s", prefix, status, task.Title)
			if idx == cursor {
				line = listSelectedStyle.Render(line)
			} else {
				line = listRowStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}
	if !focused {
		lines = append(lines, "")
		lines = append(lines, " ")
	} else {
		lines = append(lines, "")
		lines = append(lines, navActiveStyle.Render("focus: list"))
	}
	return splitRendered(listBoxStyle.Render(joinLines(lines)))
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
		return statusInboxStyle.Render(label)
	}
}
