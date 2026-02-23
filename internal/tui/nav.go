package tui

import "td/internal/domain"

type navItem struct {
	View  domain.View
	Label string
}

func defaultNavItems() []navItem {
	return []navItem{
		{View: domain.ViewToday, Label: "Today"},
		{View: domain.ViewInbox, Label: "Inbox"},
		{View: domain.ViewLog, Label: "Log"},
		{View: domain.ViewProject, Label: "Project"},
		{View: domain.ViewTrash, Label: "Trash"},
	}
}

func renderNav(items []navItem, selected int, active domain.View, focused bool) []string {
	lines := []string{"Views"}
	for idx, item := range items {
		cursor := " "
		if idx == selected {
			cursor = ">"
		}
		activeMark := " "
		if item.View == active {
			activeMark = "*"
		}
		lines = append(lines, cursor+activeMark+" "+item.Label)
	}
	if focused {
		lines = append(lines, "")
		lines = append(lines, "focus: nav")
	}
	return lines
}
