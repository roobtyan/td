package tui

import (
	"context"
	"fmt"
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
	clipUseCase  usecase.AddFromClipboardUseCase
	now          func() time.Time
	tasks        []domain.Task
	statusMsg    string
	todayDone    int
	todayTotal   int
}

func NewModel() Model {
	return NewModelWithQuery(usecase.NavQueryUseCase{})
}

func NewModelWithRepo(r repo.TaskRepository) Model {
	m := NewModelWithQuery(usecase.NewNavQueryUseCase(r))
	m.clipUseCase = usecase.AddFromClipboardUseCase{
		Repo:     r,
		AIParser: &usecase.AIParseTaskUseCase{},
	}
	return m
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
		case KeyQuit:
			return m, tea.Quit
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
		case KeyClipAdd:
			if m.clipUseCase.Repo != nil {
				_, err := m.clipUseCase.AddFromClipboard(context.Background(), "", false)
				if err != nil {
					m.statusMsg = fmt.Sprintf("clipboard add failed: %v", err)
				} else {
					m.statusMsg = "created task from clipboard"
				}
				m.reload()
			}
		case KeyClipAddAI:
			if m.clipUseCase.Repo != nil {
				_, err := m.clipUseCase.AddFromClipboard(context.Background(), "", true)
				if err != nil {
					m.statusMsg = fmt.Sprintf("ai parse failed: %v", err)
				} else {
					m.statusMsg = "created task from ai parse"
				}
				m.reload()
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	navWidth, listWidth := bodyPaneWidths(m.width)
	left := renderNav(m.navItems, m.navIndex, m.activeView, m.focus == focusNav)
	right := renderList(m.tasks, m.listCursor, m.focus == focusList)
	body := joinColumns(left, right, navWidth, listWidth)
	header := renderHeader(m.todayDone, m.todayTotal, m.activeView, len(m.tasks), m.now())
	footer := renderFooter(m.statusMsg)
	return joinSections(header, body, footer)
}

func (m *Model) reload() {
	if m.queryUseCase.Repo == nil {
		m.tasks = nil
		m.todayDone = 0
		m.todayTotal = 0
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
		m.todayDone = 0
		m.todayTotal = 0
		m.statusMsg = fmt.Sprintf("load failed: %v", err)
		return
	}
	m.tasks = tasks
	m.refreshTodayProgress()
}

func (m *Model) refreshTodayProgress() {
	if m.queryUseCase.Repo == nil {
		m.todayDone = 0
		m.todayTotal = 0
		return
	}
	all, err := m.queryUseCase.Repo.List(context.Background(), repo.TaskListFilter{})
	if err != nil {
		m.todayDone = 0
		m.todayTotal = 0
		return
	}
	now := m.now()
	done := 0
	total := 0
	for _, task := range all {
		if !isTodayProgressTask(task, now) {
			continue
		}
		total++
		if task.Status == domain.StatusDone {
			done++
		}
	}
	m.todayDone = done
	m.todayTotal = total
}

func isTodayProgressTask(task domain.Task, now time.Time) bool {
	switch task.Status {
	case domain.StatusDoing:
		return true
	case domain.StatusTodo, domain.StatusDone:
		if task.DueAt == nil {
			return false
		}
		due := task.DueAt.UTC()
		dayStart := startOfDay(now.UTC())
		dayEnd := dayStart.Add(24 * time.Hour)
		if due.Before(dayStart) {
			return true
		}
		return !due.Before(dayStart) && due.Before(dayEnd)
	default:
		return false
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
