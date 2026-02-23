package tui

import "td/internal/domain"

type navItem struct {
	View  domain.View
	Label string
}

type navRowKind int

const (
	navRowView navRowKind = iota
	navRowProject
)

type navRow struct {
	Kind    navRowKind
	View    domain.View
	Label   string
	Project string
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

func buildNavRows(items []navItem, projects []string) []navRow {
	rows := make([]navRow, 0, len(items)+len(projects))
	for _, item := range items {
		rows = append(rows, navRow{
			Kind:  navRowView,
			View:  item.View,
			Label: item.Label,
		})
		if item.View == domain.ViewProject {
			for _, project := range projects {
				rows = append(rows, navRow{
					Kind:    navRowProject,
					View:    domain.ViewProject,
					Label:   project,
					Project: project,
				})
			}
		}
	}
	return rows
}

func renderNav(rows []navRow, selected int, activeView domain.View, activeProject string, focused bool, width, height int) []string {
	contentWidth := paneContentWidth(width)
	contentHeight := paneContentHeight(height)
	lines := []string{truncateLineForPane(navTitleStyle.Render("Views"), contentWidth)}
	if contentHeight > 1 {
		visible := contentHeight - 1
		start, end := viewportWindow(len(rows), selected, visible)
		for idx := start; idx < end; idx++ {
			row := rows[idx]
			cursor := "  "
			if idx == selected {
				cursor = "> "
			}
			activeMark := " "
			if row.Kind == navRowProject {
				if activeView == domain.ViewProject && row.Project == activeProject {
					activeMark = "*"
				}
			} else if row.View == activeView && !(activeView == domain.ViewProject && activeProject != "") {
				activeMark = "*"
			}
			label := row.Label
			if row.Kind == navRowProject {
				label = "  " + label
			}
			line := truncateLineForPane(cursor+" "+activeMark+" "+label, contentWidth)
			if idx == selected {
				line = navSelectedStyle.Render(line)
			} else if activeMark == "*" {
				line = navActiveStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}
	style := navBoxStyle.Copy()
	if focused {
		style = style.BorderForeground(focusColor)
	}
	return splitRendered(renderBox(style, joinLines(lines), width, height))
}
