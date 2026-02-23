package tui

import (
	"fmt"

	"td/internal/domain"
)

func renderList(tasks []domain.Task, cursor int, focused bool) []string {
	lines := []string{"Tasks"}
	if len(tasks) == 0 {
		lines = append(lines, "(empty)")
	} else {
		for idx, task := range tasks {
			prefix := " "
			if idx == cursor {
				prefix = ">"
			}
			lines = append(lines, fmt.Sprintf("%s [%s] %s", prefix, task.Status, task.Title))
		}
	}
	if focused {
		lines = append(lines, "")
		lines = append(lines, "focus: list")
	}
	return lines
}
