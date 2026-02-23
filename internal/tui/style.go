package tui

import "github.com/charmbracelet/lipgloss"

var (
	pageBackground = lipgloss.Color("#101923")
	panelBorder    = lipgloss.Color("#2A3B4D")
	focusColor     = lipgloss.Color("#4FD1C5")
	accentColor    = lipgloss.Color("#7DD3FC")
	warnColor      = lipgloss.Color("#FBBF24")
	okColor        = lipgloss.Color("#34D399")
	mutedColor     = lipgloss.Color("#93A1B0")
)

var (
	headerBoxStyle = lipgloss.NewStyle().
			Background(pageBackground).
			Foreground(lipgloss.Color("#D9E2EC")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)

	headerInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D9E2EC")).
			Bold(true)

	logoStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	navBoxStyle = lipgloss.NewStyle().
			Background(pageBackground).
			Border(lipgloss.NormalBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)

	listBoxStyle = lipgloss.NewStyle().
			Background(pageBackground).
			Border(lipgloss.NormalBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)

	footerBoxStyle = lipgloss.NewStyle().
			Background(pageBackground).
			Foreground(mutedColor).
			Border(lipgloss.NormalBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)

	footerStatusStyle = lipgloss.NewStyle().
				Foreground(okColor)

	footerHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	navTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	navSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(focusColor).
				Bold(true)

	navActiveStyle = lipgloss.NewStyle().
			Foreground(okColor)

	listTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	listRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	listSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(lipgloss.Color("#1F3142")).
				Bold(true)

	statusInboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))
	statusTodoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
	statusDoingStyle = lipgloss.NewStyle().Foreground(warnColor)
	statusDoneStyle  = lipgloss.NewStyle().Foreground(okColor)
	statusDelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171"))
)
