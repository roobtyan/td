package tui

import "github.com/charmbracelet/lipgloss"

var (
	pageBackground  = lipgloss.Color("#101923")
	panelBorder     = lipgloss.Color("#2A3B4D")
	focusColor      = lipgloss.Color("#3FB8B3")
	accentColor     = lipgloss.Color("#96B9D8")
	warnColor       = lipgloss.Color("#FBBF24")
	okColor         = lipgloss.Color("#34D399")
	mutedColor      = lipgloss.Color("#93A1B0")
	headerTextColor = lipgloss.Color("#D9E2EC")
	logoTextColor   = headerTextColor
)

var (
	headerBoxStyle = lipgloss.NewStyle().
			Background(pageBackground).
			Foreground(headerTextColor).
			Border(lipgloss.NormalBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)

	headerInfoStyle = lipgloss.NewStyle().
			Foreground(headerTextColor).
			Bold(true)

	logoStyle = lipgloss.NewStyle().
			Foreground(logoTextColor).
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

	footerFocusStyle = lipgloss.NewStyle().
				Foreground(headerTextColor)

	footerHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	footerTokenStyle = lipgloss.NewStyle().
				Foreground(okColor).
				Bold(true)

	navTitleStyle = lipgloss.NewStyle().
			Foreground(headerTextColor).
			Bold(true)

	navSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(focusColor).
				Bold(true)

	navActiveStyle = lipgloss.NewStyle().
			Foreground(okColor)

	listTitleStyle = lipgloss.NewStyle().
			Foreground(headerTextColor).
			Bold(true)

	listRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))

	listSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8FAFC")).
				Background(lipgloss.Color("#1F3142")).
				Bold(true)

	helpBackdropLineStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Background(pageBackground)

	helpModalBoxStyle = lipgloss.NewStyle().
				Background(pageBackground).
				Foreground(headerTextColor).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(focusColor).
				Padding(0, 1)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(okColor).
			Bold(true)

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(headerTextColor).
			Bold(true)

	helpHintStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	statusInboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1"))
	statusTodoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
	statusDoingStyle = lipgloss.NewStyle().Foreground(warnColor)
	statusDoneStyle  = lipgloss.NewStyle().Foreground(okColor)
	statusDelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171"))

	metaProjectStyle    = lipgloss.NewStyle().Foreground(accentColor)
	metaDueStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC"))
	metaDueOverdueStyle = lipgloss.NewStyle().Foreground(warnColor).Bold(true)
	metaDoneStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD"))
	metaMutedStyle      = lipgloss.NewStyle().Foreground(mutedColor)

	priorityP1Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FB7185")).Bold(true)
	priorityP2Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)
	priorityP3Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
	priorityP4Style = lipgloss.NewStyle().Foreground(mutedColor)
)
