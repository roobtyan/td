package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"td/internal/app/usecase"
	"td/internal/clipboard"
	"td/internal/domain"
	"td/internal/repo"
)

type focusArea int

const (
	focusNav focusArea = iota
	focusList
)

type inputMode int

const (
	inputNone inputMode = iota
	inputAdd
	inputEdit
	inputTaskProject
	inputDue
	inputProjectCreate
	inputProjectRename
)

type undoKind int

const (
	undoTaskDelete undoKind = iota
	undoProjectDelete
	undoTaskStatus
)

type taskStatusChange struct {
	taskID int64
	from   domain.Status
	to     domain.Status
}

type undoAction struct {
	kind          undoKind
	taskIDs       []int64
	project       string
	statusChanges []taskStatusChange
}

type Model struct {
	navItems           []navItem
	navIndex           int
	activeView         domain.View
	project            string
	focus              focusArea
	listCursor         int
	width              int
	height             int
	queryUseCase       usecase.NavQueryUseCase
	clipUseCase        usecase.AddFromClipboardUseCase
	now                func() time.Time
	tasks              []domain.Task
	statusMsg          string
	todayDone          int
	todayTotal         int
	metricTodo         int
	metricDoing        int
	metricDone         int
	metricOver         int
	inputMode          inputMode
	inputValue         string
	inputCursor        int
	inputTarget        string
	projectSelectMode  bool
	projectOptions     []string
	projectSelectIndex int
	projects           []string
	showDone           bool
	showHelp           bool
	showAIInput        bool
	showAIPreview      bool
	aiInputValue       string
	aiInputCursor      int
	aiPreview          clipboard.ParsedTask
	aiPreviewRaw       string
	aiSource           string
	undoStack          []undoAction
}

const projectNoneOption = "[none]"

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

func (m Model) WithAIParser(parser *usecase.AIParseTaskUseCase) Model {
	m.clipUseCase.AIParser = parser
	return m
}

func NewModelWithQuery(uc usecase.NavQueryUseCase) Model {
	m := Model{
		navItems:     defaultNavItems(),
		navIndex:     1,
		activeView:   domain.ViewInbox,
		focus:        focusNav,
		width:        80,
		height:       24,
		queryUseCase: uc,
		now:          func() time.Time { return time.Now().Local() },
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
		m.height = msg.Height
	case tea.KeyMsg:
		if m.showHelp {
			switch msg.String() {
			case KeyHelp, KeyEsc, KeyQuit:
				m.showHelp = false
			}
			return m, nil
		}
		if m.showAIPreview {
			m.handleAIPreviewKey(msg)
			return m, nil
		}
		if m.showAIInput {
			m.handleAIInputKey(msg)
			return m, nil
		}
		if m.inputMode != inputNone {
			m.handleInputKey(msg)
			return m, nil
		}
		switch msg.String() {
		case KeyHelp:
			m.showHelp = true
		case KeyAISpace, " ":
			m.beginAIInput()
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
				rows := m.navRows()
				if m.navIndex < len(rows)-1 {
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
				row, ok := m.currentNavRow()
				if !ok {
					return m, nil
				}
				if row.Kind == navRowProject {
					m.activeView = domain.ViewProject
					m.project = row.Project
				} else {
					m.activeView = row.View
					if row.View == domain.ViewProject && m.project == "" && len(m.projects) > 0 {
						m.project = m.projects[0]
					}
				}
				m.listCursor = 0
				m.reload()
			}
		case KeyToggleDone:
			if m.activeView == domain.ViewProject {
				m.showDone = !m.showDone
				if m.showDone {
					m.statusMsg = "project: showing done tasks"
				} else {
					m.statusMsg = "project: hiding done tasks"
				}
				m.reload()
			}
		case KeyClipAdd:
			if m.clipUseCase.Repo != nil {
				_, err := m.clipUseCase.AddFromClipboard(context.Background(), "", true)
				if err != nil {
					m.statusMsg = fmt.Sprintf("ai parse failed: %v", err)
				} else {
					m.statusMsg = "created task from ai parse"
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
		case KeyAdd:
			if m.tryBeginProjectAddFromNav() {
				return m, nil
			}
			m.beginInput(inputAdd, "", "")
		case KeyEdit:
			if m.tryBeginProjectRenameFromNav() {
				return m, nil
			}
			if task, ok := m.currentTaskForAction(); ok {
				m.beginInput(inputEdit, task.Title, "")
			}
		case KeyProject:
			if task, ok := m.currentTaskForAction(); ok {
				m.beginTaskProjectInput(task.Project)
			}
		case KeyDue:
			if task, ok := m.currentTaskForAction(); ok {
				initial := ""
				if task.DueAt != nil {
					initial = formatDue(task.DueAt, m.now().Location())
				}
				m.beginInput(inputDue, initial, "")
			}
		case KeyDelete:
			if m.tryDeleteProjectFromNav() {
				return m, nil
			}
			m.removeCurrentTask()
		case KeyUndo:
			m.undoLastDelete()
		case KeyComplete:
			if m.focus == focusNav {
				if m.tryCompleteProjectFromNav() {
					return m, nil
				}
				return m, nil
			}
			m.completeCurrentTask()
		case KeyRestore:
			if m.activeView == domain.ViewTrash {
				m.restoreCurrentTrashTask()
			}
		case KeyPurgeTrash:
			if m.activeView == domain.ViewTrash {
				m.purgeTrashAll()
			}
		case KeyToday:
			m.markCurrentTaskToday()
		}
	}
	return m, nil
}

func (m Model) View() string {
	statusLine := m.statusMsg
	if m.inputMode != inputNone {
		statusLine = m.inputPrompt()
	}
	header := renderHeader(
		m.todayDone,
		m.todayTotal,
		m.activeView,
		len(m.tasks),
		m.metricDoing,
		m.metricTodo,
		m.metricDone,
		m.metricOver,
		m.now(),
		m.width,
	)
	footer := renderFooter(statusLine, m.focus, m.width, m.activeView)
	if m.inputMode != inputNone {
		footer = renderInputFooter(statusLine, m.width)
	}
	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	navRows := m.navRows()
	m.clampNavIndex()
	navWidth, listWidth, gap := bodyPaneWidths(m.width)
	left := renderNav(navRows, m.navIndex, m.activeView, m.project, m.focus == focusNav, navWidth, bodyHeight)
	right := renderList(m.tasks, m.listCursor, m.focus == focusList, listWidth, bodyHeight, m.activeView, m.now().Location())
	left = fitPaneHeight(left, bodyHeight)
	right = fitPaneHeight(right, bodyHeight)
	body := joinColumns(left, right, navWidth, listWidth, gap)
	page := joinSections(header, body, footer)
	page = fitViewport(page, m.width, m.height)
	if m.showHelp {
		dimmed := renderDimmedPage(page, m.width, m.height)
		modal := renderHelpModal(m.width)
		return overlayCentered(dimmed, modal, m.width, m.height)
	}
	if m.showAIInput {
		dimmed := renderDimmedPage(page, m.width, m.height)
		modal := renderAIInputModal(m.width, m.aiInputValue, m.aiInputCursor)
		return overlayCentered(dimmed, modal, m.width, m.height)
	}
	if m.showAIPreview {
		dimmed := renderDimmedPage(page, m.width, m.height)
		modal := renderAIPreviewModal(m.width, m.aiPreview, m.aiSource)
		return overlayCentered(dimmed, modal, m.width, m.height)
	}
	return page
}

func (m *Model) reload() {
	if m.queryUseCase.Repo == nil {
		m.tasks = nil
		m.projects = nil
		m.todayDone = 0
		m.todayTotal = 0
		return
	}
	m.refreshProjects()
	tasks, err := m.queryUseCase.ListByView(
		context.Background(),
		m.activeView,
		m.now(),
		m.project,
		m.activeView == domain.ViewProject && m.showDone,
	)
	if err != nil {
		m.tasks = nil
		m.todayDone = 0
		m.todayTotal = 0
		m.statusMsg = fmt.Sprintf("load failed: %v", err)
		return
	}
	m.tasks = tasks
	if len(m.tasks) == 0 {
		m.listCursor = 0
	} else {
		if m.listCursor < 0 {
			m.listCursor = 0
		}
		if m.listCursor >= len(m.tasks) {
			m.listCursor = len(m.tasks) - 1
		}
	}
	m.refreshTodayProgress()
}

func (m *Model) refreshTodayProgress() {
	if m.queryUseCase.Repo == nil {
		m.todayDone = 0
		m.todayTotal = 0
		m.metricTodo = 0
		m.metricDoing = 0
		m.metricDone = 0
		m.metricOver = 0
		return
	}
	all, err := m.queryUseCase.Repo.List(context.Background(), repo.TaskListFilter{})
	if err != nil {
		m.todayDone = 0
		m.todayTotal = 0
		m.metricTodo = 0
		m.metricDoing = 0
		m.metricDone = 0
		m.metricOver = 0
		return
	}
	now := m.now()
	loc := now.Location()
	done := 0
	total := 0
	todoCount := 0
	doingCount := 0
	doneCount := 0
	overdueCount := 0
	for _, task := range all {
		switch task.Status {
		case domain.StatusTodo:
			todoCount++
		case domain.StatusDoing:
			doingCount++
		case domain.StatusDone:
			doneCount++
		}
		if task.DueAt != nil {
			due := task.DueAt.In(loc)
			if due.Before(now) && (task.Status == domain.StatusInbox || task.Status == domain.StatusTodo || task.Status == domain.StatusDoing) {
				overdueCount++
			}
		}
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
	m.metricTodo = todoCount
	m.metricDoing = doingCount
	m.metricDone = doneCount
	m.metricOver = overdueCount
}

func isTodayProgressTask(task domain.Task, now time.Time) bool {
	loc := now.Location()
	dayStart := startOfDay(now)
	dayEnd := dayStart.Add(24 * time.Hour)

	switch task.Status {
	case domain.StatusDoing:
		return true
	case domain.StatusDone:
		if task.DoneAt != nil {
			done := task.DoneAt.In(loc)
			if !done.Before(dayStart) && done.Before(dayEnd) {
				return true
			}
		}
		fallthrough
	case domain.StatusTodo:
		if task.DueAt == nil {
			return false
		}
		due := task.DueAt.In(loc)
		if due.Before(dayStart) {
			return true
		}
		return !due.Before(dayStart) && due.Before(dayEnd)
	default:
		return false
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func (m *Model) refreshProjects() {
	if m.queryUseCase.Repo == nil {
		m.projects = nil
		return
	}
	uc := usecase.ProjectUseCase{Repo: m.queryUseCase.Repo}
	projects, err := uc.List(context.Background())
	if err != nil {
		m.projects = nil
		m.statusMsg = fmt.Sprintf("load projects failed: %v", err)
		return
	}
	m.projects = projects
	if m.project != "" && !containsProjectName(projects, m.project) {
		m.project = ""
	}
	if m.activeView == domain.ViewProject && m.project == "" && len(projects) > 0 {
		m.project = projects[0]
	}
	m.clampNavIndex()
}

func (m Model) navRows() []navRow {
	return buildNavRows(m.navItems, m.projects)
}

func (m *Model) clampNavIndex() {
	rows := m.navRows()
	if len(rows) == 0 {
		m.navIndex = 0
		return
	}
	if m.navIndex < 0 {
		m.navIndex = 0
	}
	if m.navIndex >= len(rows) {
		m.navIndex = len(rows) - 1
	}
}

func (m Model) currentNavRow() (navRow, bool) {
	rows := m.navRows()
	if len(rows) == 0 || m.navIndex < 0 || m.navIndex >= len(rows) {
		return navRow{}, false
	}
	return rows[m.navIndex], true
}

func (m Model) inProjectNavContext() bool {
	if m.focus != focusNav {
		return false
	}
	row, ok := m.currentNavRow()
	return ok && row.View == domain.ViewProject
}

func (m *Model) tryBeginProjectAddFromNav() bool {
	if !m.inProjectNavContext() {
		return false
	}
	m.beginInput(inputProjectCreate, "", "")
	return true
}

func (m *Model) tryBeginProjectRenameFromNav() bool {
	if !m.inProjectNavContext() {
		return false
	}
	row, ok := m.currentNavRow()
	if !ok || row.Kind != navRowProject {
		m.statusMsg = "select a project"
		return true
	}
	m.beginInput(inputProjectRename, row.Project, row.Project)
	return true
}

func (m *Model) tryDeleteProjectFromNav() bool {
	if !m.inProjectNavContext() {
		return false
	}
	row, ok := m.currentNavRow()
	if !ok || row.Kind != navRowProject {
		m.statusMsg = "select a project"
		return true
	}
	projectTasks, err := m.queryUseCase.Repo.List(context.Background(), repo.TaskListFilter{Project: row.Project})
	if err != nil {
		m.statusMsg = fmt.Sprintf("load project tasks failed: %v", err)
		return true
	}
	ids := make([]int64, 0, len(projectTasks))
	for _, task := range projectTasks {
		ids = append(ids, task.ID)
	}
	uc := usecase.ProjectUseCase{Repo: m.queryUseCase.Repo}
	if err := uc.Delete(context.Background(), row.Project); err != nil {
		m.statusMsg = fmt.Sprintf("delete project failed: %v", err)
		return true
	}
	m.pushUndo(undoAction{
		kind:    undoProjectDelete,
		project: row.Project,
		taskIDs: ids,
	})
	if m.project == row.Project {
		m.project = ""
	}
	m.statusMsg = fmt.Sprintf("deleted project %s (z undo)", row.Project)
	m.reload()
	return true
}

func (m *Model) tryCompleteProjectFromNav() bool {
	if m.focus != focusNav {
		return false
	}
	row, ok := m.currentNavRow()
	if !ok || row.View != domain.ViewProject {
		return false
	}
	if row.Kind != navRowProject {
		m.statusMsg = "select a project"
		return true
	}
	tasks, err := m.queryUseCase.Repo.List(context.Background(), repo.TaskListFilter{Project: row.Project})
	if err != nil {
		m.statusMsg = fmt.Sprintf("load project tasks failed: %v", err)
		return true
	}
	ids := make([]int64, 0, len(tasks))
	changes := make([]taskStatusChange, 0, len(tasks))
	for _, task := range tasks {
		if task.Status == domain.StatusInbox || task.Status == domain.StatusTodo || task.Status == domain.StatusDoing {
			ids = append(ids, task.ID)
			changes = append(changes, taskStatusChange{
				taskID: task.ID,
				from:   task.Status,
				to:     domain.StatusDone,
			})
		}
	}
	if len(ids) == 0 {
		m.statusMsg = fmt.Sprintf("project %s has no open task", row.Project)
		return true
	}

	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	err = uc.MarkDone(context.Background(), ids)
	if err != nil {
		m.statusMsg = fmt.Sprintf("project done failed: %v", err)
		return true
	}
	m.pushUndo(undoAction{
		kind:          undoTaskStatus,
		statusChanges: changes,
	})
	m.statusMsg = fmt.Sprintf("done %d task(s) in %s (z undo)", len(ids), row.Project)
	m.reload()
	return true
}

func (m *Model) focusProjectRow(name string) {
	if name == "" {
		return
	}
	rows := m.navRows()
	for idx, row := range rows {
		if row.Kind == navRowProject && row.Project == name {
			m.navIndex = idx
			return
		}
	}
}

func containsProjectName(projects []string, name string) bool {
	for _, item := range projects {
		if item == name {
			return true
		}
	}
	return false
}

func (m *Model) beginInput(mode inputMode, initial, target string) {
	m.inputMode = mode
	m.inputValue = initial
	m.inputCursor = len([]rune(initial))
	m.inputTarget = target
	m.projectSelectMode = false
	m.projectOptions = nil
	m.projectSelectIndex = 0
}

func (m *Model) beginTaskProjectInput(currentProject string) {
	m.beginInput(inputTaskProject, currentProject, "")
	m.projectOptions = buildProjectOptions(m.projects, currentProject)
	m.projectSelectMode = true
	m.projectSelectIndex = 0
	if currentProject != "" {
		for i, opt := range m.projectOptions {
			if opt == currentProject {
				m.projectSelectIndex = i
				break
			}
		}
	}
}

func (m *Model) endInput() {
	m.inputMode = inputNone
	m.inputValue = ""
	m.inputCursor = 0
	m.inputTarget = ""
	m.projectSelectMode = false
	m.projectOptions = nil
	m.projectSelectIndex = 0
}

func (m *Model) handleInputKey(msg tea.KeyMsg) {
	if m.inputMode == inputTaskProject && m.handleTaskProjectInputKey(msg) {
		return
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.statusMsg = "input cancelled"
		m.endInput()
	case tea.KeyEnter:
		m.submitInput()
	case tea.KeyLeft, tea.KeyCtrlB:
		m.moveInputCursor(-1)
	case tea.KeyRight, tea.KeyCtrlF:
		m.moveInputCursor(1)
	case tea.KeyBackspace, tea.KeyCtrlH:
		m.deleteInputRuneBeforeCursor()
	case tea.KeySpace:
		m.insertInputText(" ")
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			m.insertInputText(string(msg.Runes))
		}
	default:
		switch msg.String() {
		case KeyEsc:
			m.statusMsg = "input cancelled"
			m.endInput()
		case KeySelect:
			m.submitInput()
		case "left":
			m.moveInputCursor(-1)
		case "right":
			m.moveInputCursor(1)
		case KeyBackspace, KeyBackspace2:
			m.deleteInputRuneBeforeCursor()
		case " ", "space":
			m.insertInputText(" ")
		}
	}
}

func (m *Model) handleTaskProjectInputKey(msg tea.KeyMsg) bool {
	if !m.projectSelectMode {
		switch msg.Type {
		case tea.KeyTab:
			m.projectSelectMode = true
			return true
		}
		if msg.String() == KeyFocusSwitch {
			m.projectSelectMode = true
			return true
		}
		return false
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.statusMsg = "input cancelled"
		m.endInput()
		return true
	case tea.KeyEnter:
		m.submitInput()
		return true
	case tea.KeyUp:
		m.moveProjectSelection(-1)
		return true
	case tea.KeyDown:
		m.moveProjectSelection(1)
		return true
	case tea.KeyTab:
		m.projectSelectMode = false
		m.inputCursor = len([]rune(m.inputValue))
		return true
	case tea.KeyRunes:
		if len(msg.Runes) == 0 {
			return true
		}
		s := string(msg.Runes)
		if s == "j" {
			m.moveProjectSelection(1)
			return true
		}
		if s == "k" {
			m.moveProjectSelection(-1)
			return true
		}
		m.projectSelectMode = false
		m.insertInputText(s)
		return true
	}

	switch msg.String() {
	case KeyEsc:
		m.statusMsg = "input cancelled"
		m.endInput()
	case KeySelect:
		m.submitInput()
	case KeyUp, "up":
		m.moveProjectSelection(-1)
	case KeyDown, "down":
		m.moveProjectSelection(1)
	case KeyFocusSwitch:
		m.projectSelectMode = false
		m.inputCursor = len([]rune(m.inputValue))
	}
	return true
}

func (m *Model) submitInput() {
	repo := m.queryUseCase.Repo
	if repo == nil {
		m.statusMsg = "repo not ready"
		m.endInput()
		return
	}
	text := strings.TrimSpace(m.inputValue)
	focusProject := ""
	switch m.inputMode {
	case inputAdd:
		if text == "" {
			m.statusMsg = "title is empty"
			m.endInput()
			return
		}
		uc := usecase.AddTaskUseCase{Repo: repo}
		in := usecase.AddTaskInput{Title: text}
		if m.activeView == domain.ViewProject && m.project != "" {
			in.Project = m.project
		}
		task, err := uc.Execute(context.Background(), in)
		if err != nil {
			m.statusMsg = fmt.Sprintf("add failed: %v", err)
			m.endInput()
			return
		}
		if m.activeView == domain.ViewToday {
			uu := usecase.UpdateTaskUseCase{Repo: repo}
			if err := uu.MarkToday(context.Background(), []int64{task.ID}); err != nil {
				m.statusMsg = fmt.Sprintf("set today failed: %v", err)
				m.endInput()
				return
			}
		}
		m.statusMsg = fmt.Sprintf("created #%d", task.ID)
	case inputEdit:
		task, ok := m.currentTaskForAction()
		if !ok {
			m.endInput()
			return
		}
		if text == "" {
			m.statusMsg = "title is empty"
			m.endInput()
			return
		}
		uc := usecase.UpdateTaskUseCase{Repo: repo}
		if err := uc.EditTitle(context.Background(), task.ID, text); err != nil {
			m.statusMsg = fmt.Sprintf("edit failed: %v", err)
			m.endInput()
			return
		}
		m.statusMsg = fmt.Sprintf("edited #%d", task.ID)
	case inputTaskProject:
		task, ok := m.currentTaskForAction()
		if !ok {
			m.endInput()
			return
		}
		projectText := text
		if m.projectSelectMode {
			selected := m.currentProjectOption()
			if selected == projectNoneOption {
				projectText = ""
			} else {
				projectText = selected
			}
		}
		uc := usecase.UpdateTaskUseCase{Repo: repo}
		if err := uc.SetProject(context.Background(), task.ID, projectText); err != nil {
			m.statusMsg = fmt.Sprintf("set project failed: %v", err)
			m.endInput()
			return
		}
		if projectText == "" {
			m.statusMsg = fmt.Sprintf("cleared project #%d", task.ID)
		} else {
			m.statusMsg = fmt.Sprintf("project #%d -> %s", task.ID, projectText)
			focusProject = projectText
		}
	case inputDue:
		task, ok := m.currentTaskForAction()
		if !ok {
			m.endInput()
			return
		}
		var dueAt *time.Time
		if text != "" {
			due, err := parseDueInputTUI(text, m.now().Location())
			if err != nil {
				m.statusMsg = err.Error()
				return
			}
			dueAt = &due
		}
		uc := usecase.UpdateTaskUseCase{Repo: repo}
		if err := uc.SetDueAt(context.Background(), task.ID, dueAt); err != nil {
			m.statusMsg = fmt.Sprintf("set due failed: %v", err)
			m.endInput()
			return
		}
		if dueAt == nil {
			m.statusMsg = fmt.Sprintf("cleared due #%d", task.ID)
		} else {
			m.statusMsg = fmt.Sprintf("due #%d updated", task.ID)
		}
	case inputProjectCreate:
		if text == "" {
			m.statusMsg = "project name is empty"
			m.endInput()
			return
		}
		uc := usecase.ProjectUseCase{Repo: repo}
		if err := uc.Add(context.Background(), text); err != nil {
			m.statusMsg = fmt.Sprintf("create project failed: %v", err)
			m.endInput()
			return
		}
		m.activeView = domain.ViewProject
		m.project = text
		m.statusMsg = fmt.Sprintf("created project %s", text)
		focusProject = text
	case inputProjectRename:
		oldName := strings.TrimSpace(m.inputTarget)
		if oldName == "" {
			m.statusMsg = "project target is empty"
			m.endInput()
			return
		}
		if text == "" {
			m.statusMsg = "project name is empty"
			m.endInput()
			return
		}
		uc := usecase.ProjectUseCase{Repo: repo}
		if err := uc.Rename(context.Background(), oldName, text); err != nil {
			m.statusMsg = fmt.Sprintf("rename project failed: %v", err)
			m.endInput()
			return
		}
		if m.project == oldName {
			m.project = text
		}
		m.statusMsg = fmt.Sprintf("renamed project %s -> %s", oldName, text)
		focusProject = text
	}
	m.endInput()
	m.reload()
	m.focusProjectRow(focusProject)
}

func (m *Model) currentTaskForAction() (domain.Task, bool) {
	if m.focus != focusList {
		m.statusMsg = "switch focus to tasks (Tab)"
		return domain.Task{}, false
	}
	if len(m.tasks) == 0 || m.listCursor < 0 || m.listCursor >= len(m.tasks) {
		m.statusMsg = "no task selected"
		return domain.Task{}, false
	}
	return m.tasks[m.listCursor], true
}

func (m *Model) removeCurrentTask() {
	task, ok := m.currentTaskForAction()
	if !ok {
		return
	}
	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	if err := uc.Remove(context.Background(), []int64{task.ID}); err != nil {
		m.statusMsg = fmt.Sprintf("delete failed: %v", err)
		return
	}
	m.pushUndo(undoAction{
		kind:    undoTaskDelete,
		taskIDs: []int64{task.ID},
	})
	m.statusMsg = fmt.Sprintf("deleted #%d (z undo)", task.ID)
	m.reload()
	if m.listCursor >= len(m.tasks) && m.listCursor > 0 {
		m.listCursor--
	}
}

func (m *Model) restoreCurrentTrashTask() {
	task, ok := m.currentTaskForAction()
	if !ok {
		return
	}
	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	if err := uc.Restore(context.Background(), []int64{task.ID}); err != nil {
		m.statusMsg = fmt.Sprintf("restore failed: %v", err)
		return
	}
	m.statusMsg = fmt.Sprintf("restored #%d", task.ID)
	m.reload()
	if m.listCursor >= len(m.tasks) && m.listCursor > 0 {
		m.listCursor--
	}
}

func (m *Model) purgeTrashAll() {
	if len(m.tasks) == 0 {
		m.statusMsg = "trash is empty"
		return
	}
	ids := make([]int64, 0, len(m.tasks))
	for _, task := range m.tasks {
		ids = append(ids, task.ID)
	}
	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	if err := uc.Purge(context.Background(), ids); err != nil {
		m.statusMsg = fmt.Sprintf("purge failed: %v", err)
		return
	}
	m.statusMsg = fmt.Sprintf("purged %d task(s) from trash", len(ids))
	m.reload()
}

func (m *Model) pushUndo(action undoAction) {
	m.undoStack = append(m.undoStack, action)
}

func (m *Model) beginAIInput() {
	m.showAIInput = true
	m.showAIPreview = false
	m.aiInputValue = ""
	m.aiInputCursor = 0
	m.aiPreview = clipboard.ParsedTask{}
	m.aiPreviewRaw = ""
	m.aiSource = ""
}

func (m *Model) closeAIInput(reason string) {
	m.showAIInput = false
	if strings.TrimSpace(reason) != "" {
		m.statusMsg = reason
	}
}

func (m *Model) closeAIPreview(reason string) {
	m.showAIPreview = false
	if strings.TrimSpace(reason) != "" {
		m.statusMsg = reason
	}
}

func (m *Model) handleAIInputKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.closeAIInput("ai input cancelled")
	case tea.KeyEnter:
		m.submitAIInputPreview()
	case tea.KeyLeft, tea.KeyCtrlB:
		m.moveAIInputCursor(-1)
	case tea.KeyRight, tea.KeyCtrlF:
		m.moveAIInputCursor(1)
	case tea.KeyBackspace, tea.KeyCtrlH:
		m.deleteAIInputRuneBeforeCursor()
	case tea.KeySpace:
		m.insertAIInputText(" ")
	case tea.KeyRunes:
		if len(msg.Runes) > 0 {
			m.insertAIInputText(string(msg.Runes))
		}
	default:
		switch msg.String() {
		case KeyEsc, KeyQuit:
			m.closeAIInput("ai input cancelled")
		case KeySelect:
			m.submitAIInputPreview()
		case "left":
			m.moveAIInputCursor(-1)
		case "right":
			m.moveAIInputCursor(1)
		case KeyBackspace, KeyBackspace2:
			m.deleteAIInputRuneBeforeCursor()
		case " ", "space":
			m.insertAIInputText(" ")
		}
	}
}

func (m *Model) submitAIInputPreview() {
	if m.clipUseCase.Repo == nil {
		m.closeAIInput("repo not ready")
		return
	}
	text := strings.TrimSpace(m.aiInputValue)
	if text == "" {
		m.statusMsg = "ai text is empty"
		return
	}
	parsed, source, err := m.clipUseCase.ParseInput(context.Background(), text, true)
	if err != nil {
		m.closeAIInput(fmt.Sprintf("ai parse failed: %v", err))
		return
	}
	m.showAIInput = false
	m.showAIPreview = true
	m.aiPreview = parsed
	m.aiPreviewRaw = text
	m.aiSource = source
}

func (m *Model) handleAIPreviewKey(msg tea.KeyMsg) {
	switch msg.String() {
	case KeyEsc, KeyQuit:
		m.closeAIPreview("ai preview cancelled")
	case "e", "E":
		m.showAIPreview = false
		m.showAIInput = true
		m.aiInputValue = m.aiPreviewRaw
		m.aiInputCursor = len([]rune(m.aiInputValue))
	case KeySelect:
		m.confirmAIPreviewCreate()
	}
}

func (m *Model) confirmAIPreviewCreate() {
	if m.clipUseCase.Repo == nil {
		m.closeAIPreview("repo not ready")
		return
	}
	task, err := m.clipUseCase.CreateFromParsed(context.Background(), m.aiPreview)
	if err != nil {
		m.statusMsg = fmt.Sprintf("create failed: %v", err)
		return
	}
	m.showAIPreview = false
	if m.aiSource == "ai" {
		m.statusMsg = fmt.Sprintf("created #%d from ai parse", task.ID)
	} else {
		m.statusMsg = fmt.Sprintf("created #%d from fallback parse", task.ID)
	}
	m.reload()
}

func (m *Model) undoLastDelete() {
	if m.queryUseCase.Repo == nil {
		m.statusMsg = "repo not ready"
		return
	}
	if len(m.undoStack) == 0 {
		m.statusMsg = "nothing to undo"
		return
	}
	idx := len(m.undoStack) - 1
	action := m.undoStack[idx]
	ctx := context.Background()

	switch action.kind {
	case undoTaskDelete:
		uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
		if err := uc.Restore(ctx, action.taskIDs); err != nil {
			m.statusMsg = fmt.Sprintf("undo failed: %v", err)
			return
		}
		m.undoStack = m.undoStack[:idx]
		m.statusMsg = fmt.Sprintf("undid delete of %d task(s)", len(action.taskIDs))
	case undoProjectDelete:
		projectUC := usecase.ProjectUseCase{Repo: m.queryUseCase.Repo}
		if err := projectUC.Add(ctx, action.project); err != nil {
			m.statusMsg = fmt.Sprintf("undo failed: %v", err)
			return
		}
		taskUC := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
		for _, id := range action.taskIDs {
			if err := taskUC.SetProject(ctx, id, action.project); err != nil {
				m.statusMsg = fmt.Sprintf("undo failed: %v", err)
				return
			}
		}
		m.undoStack = m.undoStack[:idx]
		m.statusMsg = fmt.Sprintf("undid delete project %s", action.project)
	case undoTaskStatus:
		uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
		for i := len(action.statusChanges) - 1; i >= 0; i-- {
			change := action.statusChanges[i]
			if err := uc.SetStatus(ctx, change.taskID, change.from); err != nil {
				m.statusMsg = fmt.Sprintf("undo failed: %v", err)
				return
			}
		}
		m.undoStack = m.undoStack[:idx]
		m.statusMsg = fmt.Sprintf("undid status change of %d task(s)", len(action.statusChanges))
	default:
		m.undoStack = m.undoStack[:idx]
		m.statusMsg = "nothing to undo"
	}

	m.reload()
	if action.kind == undoProjectDelete {
		m.focusProjectRow(action.project)
	}
}

func (m *Model) markCurrentTaskToday() {
	task, ok := m.currentTaskForAction()
	if !ok {
		return
	}
	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	target := domain.StatusDoing
	if task.Status == domain.StatusDoing {
		target = domain.StatusTodo
		if err := uc.Reopen(context.Background(), []int64{task.ID}); err != nil {
			m.statusMsg = fmt.Sprintf("set todo failed: %v", err)
			return
		}
		m.statusMsg = fmt.Sprintf("todo #%d", task.ID)
	} else {
		if err := uc.MarkToday(context.Background(), []int64{task.ID}); err != nil {
			m.statusMsg = fmt.Sprintf("set today failed: %v", err)
			return
		}
		m.statusMsg = fmt.Sprintf("today #%d", task.ID)
	}
	if task.Status != target {
		m.pushUndo(undoAction{
			kind: undoTaskStatus,
			statusChanges: []taskStatusChange{{
				taskID: task.ID,
				from:   task.Status,
				to:     target,
			}},
		})
	}
	m.reload()
}

func (m *Model) completeCurrentTask() {
	task, ok := m.currentTaskForAction()
	if !ok {
		return
	}
	if task.Status == domain.StatusDone {
		m.statusMsg = fmt.Sprintf("already done #%d", task.ID)
		return
	}
	uc := usecase.UpdateTaskUseCase{Repo: m.queryUseCase.Repo}
	if err := uc.MarkDone(context.Background(), []int64{task.ID}); err != nil {
		m.statusMsg = fmt.Sprintf("set done failed: %v", err)
		return
	}
	m.pushUndo(undoAction{
		kind: undoTaskStatus,
		statusChanges: []taskStatusChange{{
			taskID: task.ID,
			from:   task.Status,
			to:     domain.StatusDone,
		}},
	})
	m.statusMsg = fmt.Sprintf("done #%d (z undo)", task.ID)
	m.reload()
	if m.listCursor >= len(m.tasks) && m.listCursor > 0 {
		m.listCursor--
	}
}

func (m Model) inputPrompt() string {
	switch m.inputMode {
	case inputAdd:
		return "add> " + renderCursorAt(m.inputValue, m.inputCursor)
	case inputEdit:
		return "edit> " + renderCursorAt(m.inputValue, m.inputCursor)
	case inputTaskProject:
		if m.projectSelectMode {
			return "project pick(j/k,Tab new)> " + m.currentProjectOption()
		}
		return "project> " + renderCursorAt(m.inputValue, m.inputCursor)
	case inputDue:
		return "due(YYYY-MM-DD HH:MM)> " + renderCursorAt(m.inputValue, m.inputCursor)
	case inputProjectCreate:
		return "project add> " + renderCursorAt(m.inputValue, m.inputCursor)
	case inputProjectRename:
		return "project rename> " + renderCursorAt(m.inputValue, m.inputCursor)
	default:
		return m.statusMsg
	}
}

func buildProjectOptions(projects []string, current string) []string {
	opts := make([]string, 0, len(projects)+2)
	opts = append(opts, projectNoneOption)
	seen := map[string]struct{}{
		projectNoneOption: {},
	}
	for _, name := range projects {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		opts = append(opts, name)
		seen[name] = struct{}{}
	}
	current = strings.TrimSpace(current)
	if current != "" {
		if _, ok := seen[current]; !ok {
			opts = append(opts, current)
		}
	}
	return opts
}

func (m *Model) currentProjectOption() string {
	if len(m.projectOptions) == 0 {
		m.projectOptions = []string{projectNoneOption}
		m.projectSelectIndex = 0
	}
	if m.projectSelectIndex < 0 {
		m.projectSelectIndex = 0
	}
	if m.projectSelectIndex >= len(m.projectOptions) {
		m.projectSelectIndex = len(m.projectOptions) - 1
	}
	return m.projectOptions[m.projectSelectIndex]
}

func (m *Model) moveProjectSelection(delta int) {
	if len(m.projectOptions) == 0 {
		m.projectOptions = []string{projectNoneOption}
		m.projectSelectIndex = 0
		return
	}
	m.projectSelectIndex += delta
	if m.projectSelectIndex < 0 {
		m.projectSelectIndex = 0
	}
	if m.projectSelectIndex >= len(m.projectOptions) {
		m.projectSelectIndex = len(m.projectOptions) - 1
	}
}

func (m *Model) moveInputCursor(delta int) {
	runes := []rune(m.inputValue)
	m.inputCursor += delta
	if m.inputCursor < 0 {
		m.inputCursor = 0
	}
	if m.inputCursor > len(runes) {
		m.inputCursor = len(runes)
	}
}

func (m *Model) insertInputText(text string) {
	m.inputValue, m.inputCursor = insertTextAtCursor(m.inputValue, m.inputCursor, text)
}

func (m *Model) deleteInputRuneBeforeCursor() {
	m.inputValue, m.inputCursor = deleteRuneBeforeCursor(m.inputValue, m.inputCursor)
}

func (m *Model) moveAIInputCursor(delta int) {
	runes := []rune(m.aiInputValue)
	m.aiInputCursor += delta
	if m.aiInputCursor < 0 {
		m.aiInputCursor = 0
	}
	if m.aiInputCursor > len(runes) {
		m.aiInputCursor = len(runes)
	}
}

func (m *Model) insertAIInputText(text string) {
	m.aiInputValue, m.aiInputCursor = insertTextAtCursor(m.aiInputValue, m.aiInputCursor, text)
}

func (m *Model) deleteAIInputRuneBeforeCursor() {
	m.aiInputValue, m.aiInputCursor = deleteRuneBeforeCursor(m.aiInputValue, m.aiInputCursor)
}

func insertTextAtCursor(base string, cursor int, inserted string) (string, int) {
	baseRunes := []rune(base)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(baseRunes) {
		cursor = len(baseRunes)
	}
	insertRunes := []rune(inserted)
	out := make([]rune, 0, len(baseRunes)+len(insertRunes))
	out = append(out, baseRunes[:cursor]...)
	out = append(out, insertRunes...)
	out = append(out, baseRunes[cursor:]...)
	return string(out), cursor + len(insertRunes)
}

func deleteRuneBeforeCursor(base string, cursor int) (string, int) {
	baseRunes := []rune(base)
	if cursor <= 0 || len(baseRunes) == 0 {
		return base, 0
	}
	if cursor > len(baseRunes) {
		cursor = len(baseRunes)
	}
	out := make([]rune, 0, len(baseRunes)-1)
	out = append(out, baseRunes[:cursor-1]...)
	out = append(out, baseRunes[cursor:]...)
	return string(out), cursor - 1
}

func parseDueInputTUI(raw string, loc *time.Location) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, fmt.Errorf("due datetime is empty")
	}
	if loc == nil {
		loc = time.Local
	}
	if t, err := time.Parse(time.RFC3339, text); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("200601021504", text, loc); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", text, loc); err == nil {
		dayEnd := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, loc)
		return dayEnd.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid due datetime")
}
