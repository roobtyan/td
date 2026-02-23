package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"td/internal/app/usecase"
	"td/internal/domain"
	"td/internal/repo"
)

type focusArea int

const (
	focusNav focusArea = iota
	focusList
)

type Model struct {
	navItems     []navItem
	navIndex     int
	activeView   domain.View
	project      string
	focus        focusArea
	listCursor   int
	width        int
	queryUseCase usecase.NavQueryUseCase
	now          func() time.Time
	tasks        []domain.Task
}

func NewModel() Model {
	return NewModelWithQuery(usecase.NavQueryUseCase{})
}

func NewModelWithRepo(r repo.TaskRepository) Model {
	return NewModelWithQuery(usecase.NewNavQueryUseCase(r))
}

func NewModelWithQuery(uc usecase.NavQueryUseCase) Model {
	m := Model{
		navItems:     defaultNavItems(),
		navIndex:     1,
		activeView:   domain.ViewInbox,
		focus:        focusNav,
		width:        80,
		queryUseCase: uc,
		now:          func() time.Time { return time.Now().UTC() },
	}
	m.reload()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case KeyFocusSwitch:
			if m.focus == focusNav {
				m.focus = focusList
			} else {
				m.focus = focusNav
			}
		case KeyDown, "down":
			if m.focus == focusNav {
				if m.navIndex < len(m.navItems)-1 {
					m.navIndex++
				}
			} else if m.listCursor < len(m.tasks)-1 {
				m.listCursor++
			}
		case KeyUp, "up":
			if m.focus == focusNav {
				if m.navIndex > 0 {
					m.navIndex--
				}
			} else if m.listCursor > 0 {
				m.listCursor--
			}
		case KeySelect:
			if m.focus == focusNav {
				m.activeView = m.navItems[m.navIndex].View
				m.listCursor = 0
				m.reload()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	navWidth, listWidth := paneWidths(m.width)
	left := renderNav(m.navItems, m.navIndex, m.activeView, m.focus == focusNav)
	right := renderList(m.tasks, m.listCursor, m.focus == focusList)
	return joinColumns(left, right, navWidth, listWidth)
}

func (m *Model) reload() {
	if m.queryUseCase.Repo == nil {
		m.tasks = nil
		return
	}
	tasks, err := m.queryUseCase.ListByView(
		context.Background(),
		m.activeView,
		m.now(),
		m.project,
	)
	if err != nil {
		m.tasks = nil
		return
	}
	m.tasks = tasks
}
