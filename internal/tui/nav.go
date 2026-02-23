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
	lines := []string{navTitleStyle.Render("Views")}
	for idx, item := range items {
		cursor := "  "
		if idx == selected {
			cursor = "> "
		}
		activeMark := " "
		if item.View == active {
			activeMark = "*"
		}
		line := cursor + activeMark + " " + item.Label
		if idx == selected {
			line = navSelectedStyle.Render(line)
		} else if item.View == active {
			line = navActiveStyle.Render(line)
		}
		lines = append(lines, line)
	}
	if !focused {
		lines = append(lines, "")
		lines = append(lines, " ")
	} else {
		lines = append(lines, "")
		lines = append(lines, navActiveStyle.Render("focus: nav"))
	}
	return splitRendered(navBoxStyle.Render(joinLines(lines)))
}
