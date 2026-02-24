package tui

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"td/internal/app/usecase"
	"td/internal/domain"
	"td/internal/repo"
)

func TestLayoutRatio(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "Today") {
		t.Fatalf("view should contain Today by default, got: %q", view)
	}
}

func TestNewModelShouldDefaultToTodayView(t *testing.T) {
	m := NewModel()
	if m.activeView != domain.ViewToday {
		t.Fatalf("activeView = %s, want %s", m.activeView, domain.ViewToday)
	}
	if m.navIndex != 0 {
		t.Fatalf("navIndex = %d, want 0", m.navIndex)
	}
}

func TestSinglePageViewContainsHeaderFooter(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "TD") {
		t.Fatalf("view should contain TD logo, got: %q", view)
	}
	if !strings.Contains(view, "Today Progress") {
		t.Fatalf("view should contain today progress, got: %q", view)
	}
	if !strings.Contains(ansi.Strip(view), "quit q") {
		t.Fatalf("view should contain quit hint, got: %q", view)
	}
}

func TestTodayProgressUsesCurrentRuleScope(t *testing.T) {
	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.UTC)
	todayDue := now.Add(2 * time.Hour)
	todayDoneDue := now.Add(1 * time.Hour)
	futureDue := now.Add(30 * time.Hour)

	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "doing-a", Status: domain.StatusDoing},
			{ID: 2, Title: "todo-today", Status: domain.StatusTodo, DueAt: &todayDue},
			{ID: 3, Title: "done-today", Status: domain.StatusDone, DueAt: &todayDoneDue},
			{ID: 4, Title: "todo-future", Status: domain.StatusTodo, DueAt: &futureDue},
		},
	}

	m := NewModelWithRepo(r)
	m.activeView = domain.ViewToday
	m.now = func() time.Time { return now }
	m.reload()

	view := m.View()
	if !strings.Contains(view, "Today Progress") {
		t.Fatalf("view should contain progress header, got: %q", view)
	}
	if !strings.Contains(view, "1/3") {
		t.Fatalf("progress should be 1/3 in today scope, got: %q", view)
	}
}

func TestTodayProgressShouldUseLocalDayBoundary(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 2, 23, 23, 30, 0, 0, loc)
	dueTomorrowLocal := time.Date(2026, 2, 24, 0, 30, 0, 0, loc)
	task := domain.Task{
		ID:     1,
		Title:  "tomorrow task",
		Status: domain.StatusTodo,
		DueAt:  &dueTomorrowLocal,
	}

	if isTodayProgressTask(task, now) {
		t.Fatalf("task due tomorrow local should not be counted in today's progress")
	}
}

func TestTodayProgressShouldCountDoneAtTodayWithoutDue(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 2, 23, 18, 30, 0, 0, loc)
	doneAt := now.Add(-45 * time.Minute)

	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "project task", Status: domain.StatusDone, DoneAt: &doneAt},
		},
	}

	m := NewModelWithRepo(r)
	m.now = func() time.Time { return now }
	m.reload()

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "1/1") {
		t.Fatalf("today progress should include done_at-today task without due, view=%q", view)
	}
}

func TestTodayViewShouldSortByPriority(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 2, 23, 10, 0, 0, 0, loc)
	dueMorning := time.Date(2026, 2, 23, 9, 30, 0, 0, loc)
	dueNoon := time.Date(2026, 2, 23, 12, 0, 0, 0, loc)

	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task-p3", Status: domain.StatusTodo, Project: "work", Priority: "P3", DueAt: &dueNoon},
			{ID: 2, Title: "task-p1", Status: domain.StatusDoing, Project: "work", Priority: "P1"},
			{ID: 3, Title: "task-p2", Status: domain.StatusTodo, Project: "work", Priority: "P2", DueAt: &dueMorning},
		},
	}

	m := NewModelWithRepo(r)
	m.now = func() time.Time { return now }
	m.activeView = domain.ViewToday
	m.reload()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m = updated.(Model)

	view := ansi.Strip(m.View())
	idxP1 := strings.Index(view, "task-p1")
	idxP2 := strings.Index(view, "task-p2")
	idxP3 := strings.Index(view, "task-p3")
	if idxP1 < 0 || idxP2 < 0 || idxP3 < 0 {
		t.Fatalf("today view should contain all tasks, view=%q", view)
	}
	if !(idxP1 < idxP2 && idxP2 < idxP3) {
		t.Fatalf("today view should sort by priority P1->P2->P3, view=%q", view)
	}
}

func TestHeaderShouldShowBalancedMetrics(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 2, 23, 10, 0, 0, 0, loc)
	overdue := now.Add(-2 * time.Hour)
	later := now.Add(2 * time.Hour)

	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "doing", Status: domain.StatusDoing},
			{ID: 2, Title: "todo-overdue", Status: domain.StatusTodo, DueAt: &overdue},
			{ID: 3, Title: "todo-later", Status: domain.StatusTodo, DueAt: &later},
			{ID: 4, Title: "done", Status: domain.StatusDone},
		},
	}

	m := NewModelWithRepo(r)
	m.now = func() time.Time { return now }
	m.activeView = domain.ViewInbox
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 24})
	m = updated.(Model)
	m.reload()

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Doing: 1") {
		t.Fatalf("header should contain Doing metric, view=%q", view)
	}
	if !strings.Contains(view, "Todo: 2") {
		t.Fatalf("header should contain Todo metric, view=%q", view)
	}
	if !strings.Contains(view, "Done: 1") {
		t.Fatalf("header should contain Done metric, view=%q", view)
	}
	if !strings.Contains(view, "Overdue: 1") {
		t.Fatalf("header should contain Overdue metric, view=%q", view)
	}
	if !strings.Contains(view, "Time:") {
		t.Fatalf("header should contain Time line, view=%q", view)
	}
	if !strings.Contains(view, "View: Inbox  Items: 2") {
		t.Fatalf("header should contain View/Items line, view=%q", view)
	}
}

func TestInboxViewShouldShowTodoWithoutProject(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "todo-no-project", Status: domain.StatusTodo},
			{ID: 2, Title: "todo-with-project", Status: domain.StatusTodo, Project: "work"},
			{ID: 3, Title: "inbox-task", Status: domain.StatusInbox},
		},
	}

	m := NewModelWithRepo(r)
	m.activeView = domain.ViewInbox
	m.reload()

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "todo-no-project") {
		t.Fatalf("inbox view should contain todo without project, view=%q", view)
	}
	if strings.Contains(view, "todo-with-project") {
		t.Fatalf("inbox view should not contain project todo, view=%q", view)
	}
}

func TestPressPShouldUseAIParseForClipboard(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m.clipUseCase.ReadClipboard = func() (string, error) {
		return "原始剪贴板内容", nil
	}
	m.clipUseCase.AIParser = &usecase.AIParseTaskUseCase{
		Provider: fakeParseProvider{
			raw: `{"title":"ai-task-title","priority":"P2"}`,
		},
	}

	m = sendRunes(m, 'p')
	if m.statusMsg != "created task from ai parse" {
		t.Fatalf("status message = %q, want %q", m.statusMsg, "created task from ai parse")
	}
	if len(r.tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(r.tasks))
	}
	if r.tasks[0].Title != "ai-task-title" {
		t.Fatalf("created task title = %q, want %q", r.tasks[0].Title, "ai-task-title")
	}
}

func TestSpaceShouldOpenAIPreviewAndConfirmCreate(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m.clipUseCase.AIParser = &usecase.AIParseTaskUseCase{
		Provider: fakeParseProvider{
			raw: `{"title":"在线推理功能","project":"PP Polaris","priority":"P1","due":"2026-02-25 13:00"}`,
		},
	}

	m = sendRunes(m, ' ')
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "AI QUICK INPUT") {
		t.Fatalf("space should open ai input modal, view=%q", view)
	}

	m = sendText(m, "在PP Polaris项目下 明天下午一点完成在线推理功能")
	m = sendEnter(m)
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "AI PREVIEW") {
		t.Fatalf("enter on ai input should open preview modal, view=%q", view)
	}
	if !strings.Contains(view, "source") || !strings.Contains(view, "AI") {
		t.Fatalf("preview should show source AI, view=%q", view)
	}
	if !strings.Contains(view, "project") || !strings.Contains(view, "PP Polaris") {
		t.Fatalf("preview should show project, view=%q", view)
	}
	if !strings.Contains(view, "priority") || !strings.Contains(view, "P1") {
		t.Fatalf("preview should show priority, view=%q", view)
	}
	if !strings.Contains(view, "due") || !strings.Contains(view, "2026-02-25 13:00") {
		t.Fatalf("preview should show due, view=%q", view)
	}

	m = sendEnter(m)
	if len(r.tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(r.tasks))
	}
	if r.tasks[0].Project != "PP Polaris" {
		t.Fatalf("task project = %q, want %q", r.tasks[0].Project, "PP Polaris")
	}
	if r.tasks[0].Priority != "P1" {
		t.Fatalf("task priority = %q, want %q", r.tasks[0].Priority, "P1")
	}
	if r.tasks[0].DueAt == nil {
		t.Fatalf("task due should not be nil")
	}
	if !strings.Contains(m.statusMsg, "from ai parse") {
		t.Fatalf("status message = %q, want ai source hint", m.statusMsg)
	}
}

func TestFooterInputShouldShowCursor(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, 'a')
	m = sendText(m, "abc")

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "add> abc|") {
		t.Fatalf("input footer should show cursor marker, view=%q", view)
	}
}

func TestAISpaceInputModalShouldShowCursor(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, ' ')
	m = sendText(m, "abc")

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "abc|") {
		t.Fatalf("ai input modal should show cursor marker, view=%q", view)
	}
}

func TestFooterInputShouldSupportLeftRightMoveAndInsert(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, 'a')
	m = sendText(m, "abc")
	m = sendLeft(m)
	m = sendRunes(m, 'X')

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "add> abX|c") {
		t.Fatalf("input footer should support cursor move and mid insert, view=%q", view)
	}
}

func TestAISpaceInputModalShouldSupportLeftRightMoveAndInsert(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, ' ')
	m = sendText(m, "abc")
	m = sendLeft(m)
	m = sendRunes(m, 'X')

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "abX|c") {
		t.Fatalf("ai input modal should support cursor move and mid insert, view=%q", view)
	}
}

func TestFooterInputShouldSupportCtrlBFCursorMove(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, 'a')
	m = sendText(m, "abc")
	m = sendCtrlB(m)
	m = sendRunes(m, 'X')
	m = sendCtrlF(m)
	m = sendRunes(m, 'Y')

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "add> abXcY|") {
		t.Fatalf("input footer should support ctrl+b/ctrl+f cursor move, view=%q", view)
	}
}

func TestAISpaceInputModalShouldSupportCtrlBFCursorMove(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m = sendRunes(m, ' ')
	m = sendText(m, "abc")
	m = sendCtrlB(m)
	m = sendRunes(m, 'X')
	m = sendCtrlF(m)
	m = sendRunes(m, 'Y')

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "abXcY|") {
		t.Fatalf("ai input should support ctrl+b/ctrl+f cursor move, view=%q", view)
	}
}

func TestPressQShouldQuit(t *testing.T) {
	m := NewModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("q should return quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q cmd should emit tea.QuitMsg")
	}
}

func TestViewShouldFitWindowSize(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 60, Height: 12})
	m := updated.(Model)

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 12 {
		t.Fatalf("view line count = %d, want <= 12", len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		width := ansi.StringWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth > 60 {
		t.Fatalf("view max line width = %d, want <= 60", maxWidth)
	}
}

func TestViewShouldKeepFooterInSmallWindow(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 12})
	m := updated.(Model)

	view := m.View()
	if !strings.Contains(view, "ready") {
		t.Fatalf("view should keep footer section after resize, got: %q", view)
	}
}

func TestViewShouldFitMultipleWindowSizes(t *testing.T) {
	cases := []struct {
		name   string
		width  int
		height int
	}{
		{name: "large", width: 120, height: 30},
		{name: "medium", width: 80, height: 20},
		{name: "small", width: 60, height: 12},
		{name: "narrow", width: 38, height: 10},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			updated, _ := model.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			m := updated.(Model)
			assertViewFits(t, m.View(), tt.width, tt.height)
		})
	}
}

func TestHeaderAndFooterShouldUseFullWidth(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 130, Height: 24})
	m := updated.(Model)
	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 4 {
		t.Fatalf("view lines too short: %d", len(lines))
	}

	headerTop := strings.TrimRight(ansi.Strip(lines[0]), " ")
	if ansi.StringWidth(headerTop) != 130 {
		t.Fatalf("header top visible width = %d, want 130, line=%q", ansi.StringWidth(headerTop), lines[0])
	}

	footerLine := ""
	for _, line := range lines {
		if strings.Contains(line, "ready") {
			footerLine = line
			break
		}
	}
	if footerLine == "" {
		t.Fatalf("footer line not found, view=%q", view)
	}
	trimmed := strings.TrimRight(ansi.Strip(footerLine), " ")
	if ansi.StringWidth(trimmed) != 130 {
		t.Fatalf("footer visible width = %d, want 130, line=%q", ansi.StringWidth(trimmed), footerLine)
	}
}

func TestBodyShouldNotUseHardVerticalSeparator(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 24})
	m := updated.(Model)
	view := m.View()

	separatorPattern := regexp.MustCompile(`\s{6,}\|\s{6,}`)
	if separatorPattern.MatchString(view) {
		t.Fatalf("body should not contain hard separator column, view=%q", view)
	}
}

func TestFocusIndicatorShouldOnlyAppearInFooter(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m := updated.(Model)
	view := m.View()
	plain := ansi.Strip(view)

	if strings.Contains(plain, "focus: nav") || strings.Contains(plain, "focus: list") {
		t.Fatalf("focus indicator should not stay inside panes, view=%q", view)
	}
	if !strings.Contains(plain, "focus nav") {
		t.Fatalf("focus indicator should appear in footer, view=%q", view)
	}
}

func TestViewShouldNotShowTruncationArtifactsInNormalWidth(t *testing.T) {
	model := NewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m := updated.(Model)
	view := m.View()
	if strings.Contains(view, "...") {
		t.Fatalf("view should not contain truncation artifacts at normal width, view=%q", view)
	}
}

func TestFooterShouldOnlyShowInputPromptInInputMode(t *testing.T) {
	m := NewModel()
	m = sendRunes(m, 'a')
	m = sendText(m, "draft")

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "add> draft") {
		t.Fatalf("input prompt should be visible in footer, view=%q", view)
	}
	if strings.Contains(view, "quit q") || strings.Contains(view, "help ?") {
		t.Fatalf("footer shortcuts should be hidden in input mode, view=%q", view)
	}
}

func TestFooterShouldRestoreShortcutsAfterInputEnd(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)

	m = sendRunes(m, 'a')
	m = sendText(m, "one")
	m = sendEnter(m)

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "quit q") || !strings.Contains(view, "help ?") {
		t.Fatalf("footer shortcuts should be restored after input submit, view=%q", view)
	}
	if strings.Contains(view, "j/k move") {
		t.Fatalf("footer should not show full shortcut list in normal mode, view=%q", view)
	}
}

func TestTrashFooterShouldShowRestoreAndPurgeHints(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "trash a", Status: domain.StatusDeleted},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewTrash
	m.reload()

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "restore r") || !strings.Contains(view, "purge X") {
		t.Fatalf("trash footer should contain restore/purge hints, view=%q", view)
	}
}

func TestNonTrashFooterShouldHideRestoreAndPurgeHints(t *testing.T) {
	m := NewModel()
	view := ansi.Strip(m.View())
	if strings.Contains(view, "restore r") || strings.Contains(view, "purge X") {
		t.Fatalf("non-trash footer should not contain restore/purge hints, view=%q", view)
	}
}

func TestHelpModalShouldToggleByQuestionMark(t *testing.T) {
	m := NewModel()
	m = sendRunes(m, '?')

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "HELP") {
		t.Fatalf("help modal title should be visible, view=%q", view)
	}
	if !strings.Contains(view, "press ? / q / esc to close") {
		t.Fatalf("help modal close hint should be visible, view=%q", view)
	}

	m = sendRunes(m, '?')
	view = ansi.Strip(m.View())
	if strings.Contains(view, "HELP") {
		t.Fatalf("help modal should close after pressing ?, view=%q", view)
	}
}

func TestHelpModalShouldCloseByQWithoutQuit(t *testing.T) {
	m := NewModel()
	m = sendRunes(m, '?')

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatalf("pressing q in help mode should not quit app")
	}
	m = updated.(Model)
	if m.showHelp {
		t.Fatalf("help modal should be closed by q")
	}
}

func TestUIAddTaskByInputMode(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)

	m = sendRunes(m, 'a')
	m = sendText(m, "write report")
	m = sendEnter(m)

	if len(r.tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(r.tasks))
	}
	if r.tasks[0].Title != "write report" {
		t.Fatalf("title = %q, want %q", r.tasks[0].Title, "write report")
	}
}

func TestUIAddTaskInTodayShouldBeDoing(t *testing.T) {
	r := &fakeTaskRepo{}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewToday
	m.reload()

	m = sendRunes(m, 'a')
	m = sendText(m, "today task")
	m = sendEnter(m)

	if len(r.tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(r.tasks))
	}
	if r.tasks[0].Status != domain.StatusDoing {
		t.Fatalf("status = %s, want %s", r.tasks[0].Status, domain.StatusDoing)
	}
}

func TestUIEditProjectTodayDueAndDelete(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)

	m = sendTab(m)

	m = sendRunes(m, 'e')
	for i := 0; i < len("task a"); i++ {
		m = sendBackspace(m)
	}
	m = sendText(m, "task updated")
	m = sendEnter(m)
	if r.tasks[0].Title != "task updated" {
		t.Fatalf("title = %q, want %q", r.tasks[0].Title, "task updated")
	}

	m = sendRunes(m, 'P')
	m = sendText(m, "work")
	m = sendEnter(m)
	if r.tasks[0].Project != "work" {
		t.Fatalf("project = %q, want %q", r.tasks[0].Project, "work")
	}

	m = sendRunes(m, 'd')
	m = sendText(m, "2026-02-25 18:00")
	m = sendEnter(m)
	if r.tasks[0].DueAt == nil {
		t.Fatalf("due_at should not be nil")
	}

	m = sendRunes(m, 't')
	if r.tasks[0].Status != domain.StatusDoing {
		t.Fatalf("status = %s, want %s", r.tasks[0].Status, domain.StatusDoing)
	}
}

func TestEditInputShouldPrefillCurrentTaskTitle(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'e')

	if m.inputMode != inputEdit {
		t.Fatalf("inputMode = %v, want %v", m.inputMode, inputEdit)
	}
	if m.inputValue != "task a" {
		t.Fatalf("inputValue = %q, want %q", m.inputValue, "task a")
	}
	if m.inputCursor != len([]rune("task a")) {
		t.Fatalf("inputCursor = %d, want %d", m.inputCursor, len([]rune("task a")))
	}
}

func TestProjectInputShouldPrefillCurrentTaskProject(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'P')

	if m.inputMode != inputTaskProject {
		t.Fatalf("inputMode = %v, want %v", m.inputMode, inputTaskProject)
	}
	if m.inputValue != "work" {
		t.Fatalf("inputValue = %q, want %q", m.inputValue, "work")
	}
	if m.inputCursor != len([]rune("work")) {
		t.Fatalf("inputCursor = %d, want %d", m.inputCursor, len([]rune("work")))
	}
}

func TestProjectInputShouldSelectExistingProjectByJKAndEnter(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"alpha", "beta"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'P')
	m = sendRunes(m, 'j')
	m = sendEnter(m)

	if r.tasks[0].Project != "alpha" {
		t.Fatalf("project = %q, want %q", r.tasks[0].Project, "alpha")
	}
}

func TestProjectInputShouldCreateNewProjectInInputMode(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"alpha"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'P')
	m = sendTab(m)
	m = sendText(m, "newproj")
	m = sendEnter(m)

	if r.tasks[0].Project != "newproj" {
		t.Fatalf("project = %q, want %q", r.tasks[0].Project, "newproj")
	}
	if !containsString(r.projects, "newproj") {
		t.Fatalf("projects = %v, want contains newproj", r.projects)
	}
}

func TestProjectInputShouldClearProjectBySelectingNone(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"alpha"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Project: "alpha"},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewProject
	m.project = "alpha"
	m.reload()
	m = sendTab(m)
	m = sendRunes(m, 'P')
	m = sendRunes(m, 'k')
	if !m.projectSelectMode {
		t.Fatalf("projectSelectMode = false, want true")
	}
	if got := m.currentProjectOption(); got != projectNoneOption {
		t.Fatalf("selected option = %q, want %q", got, projectNoneOption)
	}
	m = sendEnter(m)

	if r.tasks[0].Project != "" {
		t.Fatalf("project = %q, want empty", r.tasks[0].Project)
	}
}

func TestDueInputShouldPrefillCurrentTaskDue(t *testing.T) {
	due := time.Date(2026, 2, 25, 18, 0, 0, 0, time.Local)
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, DueAt: &due},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'd')

	want := due.In(time.Local).Format("2006-01-02 15:04")
	if m.inputMode != inputDue {
		t.Fatalf("inputMode = %v, want %v", m.inputMode, inputDue)
	}
	if m.inputValue != want {
		t.Fatalf("inputValue = %q, want %q", m.inputValue, want)
	}
	if m.inputCursor != len([]rune(want)) {
		t.Fatalf("inputCursor = %d, want %d", m.inputCursor, len([]rune(want)))
	}
}

func TestPriorityInputShouldPrefillCurrentTaskPriority(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Priority: "P3"},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'y')

	if m.inputValue != "P3" {
		t.Fatalf("inputValue = %q, want %q", m.inputValue, "P3")
	}
	if m.inputCursor != len([]rune("P3")) {
		t.Fatalf("inputCursor = %d, want %d", m.inputCursor, len([]rune("P3")))
	}
}

func TestUIPriorityShouldUpdateByY(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Priority: "P3"},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'y')
	m = sendBackspace(m)
	m = sendBackspace(m)
	m = sendText(m, "P1")
	m = sendEnter(m)

	if r.tasks[0].Priority != "P1" {
		t.Fatalf("priority = %q, want %q", r.tasks[0].Priority, "P1")
	}
}

func TestProjectRenameShouldPrefillSelectedProjectName(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'e')

	if m.inputMode != inputProjectRename {
		t.Fatalf("inputMode = %v, want %v", m.inputMode, inputProjectRename)
	}
	if m.inputValue != "work" {
		t.Fatalf("inputValue = %q, want %q", m.inputValue, "work")
	}
	if m.inputCursor != len([]rune("work")) {
		t.Fatalf("inputCursor = %d, want %d", m.inputCursor, len([]rune("work")))
	}
}

func TestInputShouldAcceptSpaceKeyForDue(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'd')
	m = sendText(m, "2026-02-25")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(Model)
	m = sendText(m, "18:00")
	m = sendEnter(m)

	if r.tasks[0].DueAt == nil {
		t.Fatalf("due_at should not be nil after typing with space key, input=%q status=%q", m.inputValue, m.statusMsg)
	}
}

func TestInputShouldAcceptCompactDateTimeForDue(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'd')
	m = sendText(m, "202602051122")
	m = sendEnter(m)

	if r.tasks[0].DueAt == nil {
		t.Fatalf("due_at should not be nil after compact datetime input")
	}
	want := time.Date(2026, 2, 5, 11, 22, 0, 0, time.Local).UTC()
	if !r.tasks[0].DueAt.Equal(want) {
		t.Fatalf("due_at = %s, want %s", r.tasks[0].DueAt.UTC(), want)
	}
}

func TestUIDeleteTask(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'x')
	if r.tasks[0].Status != domain.StatusDeleted {
		t.Fatalf("status = %s, want %s", r.tasks[0].Status, domain.StatusDeleted)
	}
	if !strings.Contains(m.statusMsg, "z undo") {
		t.Fatalf("status message should hint z undo, got %q", m.statusMsg)
	}
}

func TestUIDeleteTaskShouldUndoByZ(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'x')
	if r.tasks[0].Status != domain.StatusDeleted {
		t.Fatalf("status after delete = %s, want %s", r.tasks[0].Status, domain.StatusDeleted)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusTodo {
		t.Fatalf("status after undo = %s, want %s", r.tasks[0].Status, domain.StatusTodo)
	}
}

func TestTrashViewShouldShowProjectInTaskRow(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "trash a", Status: domain.StatusDeleted, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewTrash
	m.reload()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m = updated.(Model)

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "work") {
		t.Fatalf("trash row should contain project name, view=%q", view)
	}
	if strings.Contains(view, "proj:") {
		t.Fatalf("trash row should not contain explicit project field label, view=%q", view)
	}
}

func TestTrashShouldRestoreSelectedByR(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "trash a", Status: domain.StatusDeleted},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewTrash
	m.reload()
	m = sendTab(m)
	m = sendRunes(m, 'r')

	if r.tasks[0].Status != domain.StatusTodo {
		t.Fatalf("status after restore = %s, want %s", r.tasks[0].Status, domain.StatusTodo)
	}
}

func TestTrashShouldPurgeAllByUpperX(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "trash a", Status: domain.StatusDeleted},
			{ID: 2, Title: "trash b", Status: domain.StatusDeleted},
			{ID: 3, Title: "todo a", Status: domain.StatusTodo},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewTrash
	m.reload()
	m = sendRunes(m, 'X')

	if len(r.tasks) != 1 {
		t.Fatalf("task count after purge = %d, want 1", len(r.tasks))
	}
	if r.tasks[0].Status != domain.StatusTodo {
		t.Fatalf("remaining task status = %s, want %s", r.tasks[0].Status, domain.StatusTodo)
	}
}

func TestProjectViewToggleDoneByH(t *testing.T) {
	now := time.Date(2026, 2, 23, 10, 0, 0, 0, time.Local)
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "todo in work", Status: domain.StatusTodo, Project: "work"},
			{ID: 2, Title: "done in work", Status: domain.StatusDone, Project: "work", DoneAt: &now},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewProject
	m.project = "work"
	m.reload()
	before := ansi.Strip(m.View())
	if strings.Contains(before, "done in work") {
		t.Fatalf("before toggle should hide done task, view=%q", before)
	}
	m = sendRunes(m, 'h')
	after := ansi.Strip(m.View())
	if !strings.Contains(after, "done in work") {
		t.Fatalf("after toggle should show done task, view=%q", after)
	}
}

func TestListShouldShowDueInTaskRow(t *testing.T) {
	due := time.Date(2026, 2, 24, 9, 30, 0, 0, time.Local)
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "with due", Status: domain.StatusInbox, Priority: "P2", DueAt: &due},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m = updated.(Model)
	view := ansi.Strip(m.View())
	taskLine := lineContaining(view, "with due")
	if taskLine == "" {
		t.Fatalf("task row should exist, view=%q", view)
	}
	if !strings.Contains(taskLine, "2026-02-24 09:30") {
		t.Fatalf("task row should contain due text, view=%q", view)
	}
	if !strings.Contains(taskLine, "P2") {
		t.Fatalf("task row should contain priority text, view=%q", view)
	}
	if strings.Contains(taskLine, "due:") {
		t.Fatalf("task row should not contain explicit due field label, view=%q", view)
	}
}

func TestTodayViewTaskRowShouldShowProjectAndDue(t *testing.T) {
	due := time.Date(2026, 2, 24, 9, 30, 0, 0, time.Local)
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "today task", Status: domain.StatusDoing, Project: "work", Priority: "P1", DueAt: &due},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewToday
	m.reload()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m = updated.(Model)

	view := ansi.Strip(m.View())
	taskLine := lineContaining(view, "today task")
	if taskLine == "" {
		t.Fatalf("today row should exist, view=%q", view)
	}
	if !strings.Contains(taskLine, "work") {
		t.Fatalf("today row should contain project text, view=%q", view)
	}
	if !strings.Contains(taskLine, "2026-02-24 09:30") {
		t.Fatalf("today row should contain due text, view=%q", view)
	}
	if !strings.Contains(taskLine, "P1") {
		t.Fatalf("today row should contain priority, view=%q", view)
	}
	if strings.Contains(taskLine, "proj:") || strings.Contains(taskLine, "due:") {
		t.Fatalf("today row should not contain explicit field labels, view=%q", view)
	}
}

func TestProjectAddKeyInNavShouldCreateProject(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'a')
	m = sendText(m, "home")
	m = sendEnter(m)
	if !containsString(r.projects, "home") {
		t.Fatalf("projects = %v, want contains home", r.projects)
	}
}

func TestProjectRenameAndDeleteByNavKeys(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'e')
	for i := 0; i < len("work"); i++ {
		m = sendBackspace(m)
	}
	m = sendText(m, "office")
	m = sendEnter(m)
	if !containsString(r.projects, "office") || containsString(r.projects, "work") {
		t.Fatalf("projects after rename = %v, want contains office and no work", r.projects)
	}
	if r.tasks[0].Project != "office" {
		t.Fatalf("task project after rename = %q, want %q", r.tasks[0].Project, "office")
	}

	m = sendRunes(m, 'x')
	if containsString(r.projects, "office") {
		t.Fatalf("projects after delete = %v, should not contain office", r.projects)
	}
	if r.tasks[0].Project != "" {
		t.Fatalf("task project after delete = %q, want empty", r.tasks[0].Project)
	}
	if !strings.Contains(m.statusMsg, "z undo") {
		t.Fatalf("status message should hint z undo, got %q", m.statusMsg)
	}
}

func TestProjectDeleteShouldUndoByZ(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'x')
	if containsString(r.projects, "work") {
		t.Fatalf("project should be deleted, projects=%v", r.projects)
	}
	if r.tasks[0].Project != "" {
		t.Fatalf("task project after delete = %q, want empty", r.tasks[0].Project)
	}

	m = sendRunes(m, 'z')
	if !containsString(r.projects, "work") {
		t.Fatalf("project should be restored, projects=%v", r.projects)
	}
	if r.tasks[0].Project != "work" {
		t.Fatalf("task project after undo = %q, want %q", r.tasks[0].Project, "work")
	}
}

func TestUndoShouldSupportMultipleStepsForTaskDelete(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
			{ID: 2, Title: "task b", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)

	m = sendRunes(m, 'x')
	m = sendRunes(m, 'x')
	if r.tasks[0].Status != domain.StatusDeleted || r.tasks[1].Status != domain.StatusDeleted {
		t.Fatalf("after double delete statuses = %s/%s, want deleted/deleted", r.tasks[0].Status, r.tasks[1].Status)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusDeleted || r.tasks[1].Status != domain.StatusTodo {
		t.Fatalf("after first undo statuses = %s/%s, want deleted/todo", r.tasks[0].Status, r.tasks[1].Status)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusTodo || r.tasks[1].Status != domain.StatusTodo {
		t.Fatalf("after second undo statuses = %s/%s, want todo/todo", r.tasks[0].Status, r.tasks[1].Status)
	}
}

func TestLogViewShouldShowProjectAndDoneAt(t *testing.T) {
	now := time.Date(2026, 2, 23, 20, 0, 0, 0, time.Local)
	doneAt := now.Add(-30 * time.Minute)
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "测试用例", Status: domain.StatusDone, Project: "xxx", Priority: "P2", DoneAt: &doneAt},
		},
	}
	m := NewModelWithRepo(r)
	m.now = func() time.Time { return now }
	m.activeView = domain.ViewLog
	m.reload()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m = updated.(Model)

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "xxx") {
		t.Fatalf("log row should contain project, view=%q", view)
	}
	if !strings.Contains(view, "2026-02-23 19:30") {
		t.Fatalf("log row should contain done time, view=%q", view)
	}
	if !strings.Contains(view, "P2") {
		t.Fatalf("log row should contain priority, view=%q", view)
	}
	if strings.Contains(view, "proj:") || strings.Contains(view, "done:") {
		t.Fatalf("log row should not contain explicit field labels, view=%q", view)
	}
}

func TestTaskShouldMarkDoneByCInListFocus(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)
	m = sendRunes(m, 'c')
	if r.tasks[0].Status != domain.StatusDone {
		t.Fatalf("status = %s, want %s", r.tasks[0].Status, domain.StatusDone)
	}
}

func TestTodayToggleByTShouldSwitchDoingAndTodo(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "doing task", Status: domain.StatusDoing, Project: "work"},
			{ID: 2, Title: "todo task", Status: domain.StatusTodo, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewProject
	m.project = "work"
	m.reload()
	m = sendTab(m)

	m = sendRunes(m, 't')
	if r.tasks[0].Status != domain.StatusTodo {
		t.Fatalf("doing task status after t = %s, want %s, statusMsg=%q focus=%v listLen=%d", r.tasks[0].Status, domain.StatusTodo, m.statusMsg, m.focus, len(m.tasks))
	}

	m = sendRunes(m, 'j')
	m = sendRunes(m, 't')
	if r.tasks[1].Status != domain.StatusDoing {
		t.Fatalf("todo task status after t = %s, want %s, statusMsg=%q focus=%v listLen=%d", r.tasks[1].Status, domain.StatusDoing, m.statusMsg, m.focus, len(m.tasks))
	}
}

func TestProjectShouldMarkAllOpenTasksDoneByCInNav(t *testing.T) {
	now := time.Date(2026, 2, 23, 20, 0, 0, 0, time.Local)
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "a", Status: domain.StatusInbox, Project: "work"},
			{ID: 2, Title: "b", Status: domain.StatusTodo, Project: "work"},
			{ID: 3, Title: "c", Status: domain.StatusDoing, Project: "work"},
			{ID: 4, Title: "d", Status: domain.StatusDone, Project: "work", DoneAt: &now},
			{ID: 5, Title: "e", Status: domain.StatusTodo, Project: "home"},
		},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'c')

	if r.tasks[0].Status != domain.StatusDone || r.tasks[1].Status != domain.StatusDone || r.tasks[2].Status != domain.StatusDone {
		t.Fatalf("work open tasks should be done, got statuses: %s %s %s", r.tasks[0].Status, r.tasks[1].Status, r.tasks[2].Status)
	}
	if r.tasks[3].Status != domain.StatusDone {
		t.Fatalf("already done task should stay done, got %s", r.tasks[3].Status)
	}
	if r.tasks[4].Status != domain.StatusTodo {
		t.Fatalf("other project task should stay todo, got %s", r.tasks[4].Status)
	}
}

func TestTaskStatusShouldUndoTodoDoingDoneChainByZ(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusTodo, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.activeView = domain.ViewProject
	m.project = "work"
	m.reload()
	m = sendTab(m)

	m = sendRunes(m, 't')
	if r.tasks[0].Status != domain.StatusDoing {
		t.Fatalf("status after t = %s, want %s", r.tasks[0].Status, domain.StatusDoing)
	}

	m = sendRunes(m, 'c')
	if r.tasks[0].Status != domain.StatusDone {
		t.Fatalf("status after c = %s, want %s", r.tasks[0].Status, domain.StatusDone)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusDoing {
		t.Fatalf("status after first z = %s, want %s", r.tasks[0].Status, domain.StatusDoing)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusTodo {
		t.Fatalf("status after second z = %s, want %s", r.tasks[0].Status, domain.StatusTodo)
	}
}

func TestTaskStatusUndoShouldRestoreInboxFromDoneByZ(t *testing.T) {
	r := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: 1, Title: "task a", Status: domain.StatusInbox},
		},
	}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	m = sendTab(m)

	m = sendRunes(m, 'c')
	if r.tasks[0].Status != domain.StatusDone {
		t.Fatalf("status after c = %s, want %s", r.tasks[0].Status, domain.StatusDone)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusInbox {
		t.Fatalf("status after z = %s, want %s", r.tasks[0].Status, domain.StatusInbox)
	}
}

func TestProjectDoneByCInNavShouldUndoByZ(t *testing.T) {
	now := time.Date(2026, 2, 23, 20, 0, 0, 0, time.Local)
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "a", Status: domain.StatusInbox, Project: "work"},
			{ID: 2, Title: "b", Status: domain.StatusTodo, Project: "work"},
			{ID: 3, Title: "c", Status: domain.StatusDoing, Project: "work"},
			{ID: 4, Title: "d", Status: domain.StatusDone, Project: "work", DoneAt: &now},
			{ID: 5, Title: "e", Status: domain.StatusTodo, Project: "home"},
		},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'c')

	if r.tasks[0].Status != domain.StatusDone || r.tasks[1].Status != domain.StatusDone || r.tasks[2].Status != domain.StatusDone {
		t.Fatalf("work open tasks should be done before undo")
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusInbox {
		t.Fatalf("task1 status after z = %s, want %s", r.tasks[0].Status, domain.StatusInbox)
	}
	if r.tasks[1].Status != domain.StatusTodo {
		t.Fatalf("task2 status after z = %s, want %s", r.tasks[1].Status, domain.StatusTodo)
	}
	if r.tasks[2].Status != domain.StatusDoing {
		t.Fatalf("task3 status after z = %s, want %s", r.tasks[2].Status, domain.StatusDoing)
	}
	if r.tasks[3].Status != domain.StatusDone {
		t.Fatalf("task4 status after z = %s, want %s", r.tasks[3].Status, domain.StatusDone)
	}
}

func TestUndoShouldUseGlobalLatestAcrossProjectStatusAndDelete(t *testing.T) {
	r := &fakeTaskRepo{
		projects: []string{"work"},
		tasks: []domain.Task{
			{ID: 1, Title: "a", Status: domain.StatusTodo, Project: "work"},
			{ID: 2, Title: "b", Status: domain.StatusDoing, Project: "work"},
		},
	}
	m := NewModelWithRepo(r)
	m.navIndex = 1
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')
	m = sendRunes(m, 'j')

	m = sendRunes(m, 'c')
	if r.tasks[0].Status != domain.StatusDone || r.tasks[1].Status != domain.StatusDone {
		t.Fatalf("statuses after c = %s/%s, want done/done", r.tasks[0].Status, r.tasks[1].Status)
	}

	m = sendRunes(m, 'x')
	if containsString(r.projects, "work") {
		t.Fatalf("project should be deleted")
	}
	if r.tasks[0].Project != "" || r.tasks[1].Project != "" {
		t.Fatalf("project field should be detached after delete, got %q/%q", r.tasks[0].Project, r.tasks[1].Project)
	}

	m = sendRunes(m, 'z')
	if !containsString(r.projects, "work") {
		t.Fatalf("first z should undo delete project")
	}
	if r.tasks[0].Project != "work" || r.tasks[1].Project != "work" {
		t.Fatalf("first z should restore project relation, got %q/%q", r.tasks[0].Project, r.tasks[1].Project)
	}
	if r.tasks[0].Status != domain.StatusDone || r.tasks[1].Status != domain.StatusDone {
		t.Fatalf("first z should not undo status yet, got %s/%s", r.tasks[0].Status, r.tasks[1].Status)
	}

	m = sendRunes(m, 'z')
	if r.tasks[0].Status != domain.StatusTodo || r.tasks[1].Status != domain.StatusDoing {
		t.Fatalf("second z should undo project status operation, got %s/%s", r.tasks[0].Status, r.tasks[1].Status)
	}
}

func TestLongTaskListShouldKeepFooterAndSelectedVisible(t *testing.T) {
	tasks := make([]domain.Task, 0, 80)
	for i := 1; i <= 80; i++ {
		tasks = append(tasks, domain.Task{
			ID:     int64(i),
			Title:  "task-" + fmt.Sprintf("%02d", i) + strings.Repeat("-very-long-title", 3),
			Status: domain.StatusInbox,
		})
	}
	r := &fakeTaskRepo{tasks: tasks}
	m := NewModelWithRepo(r)
	m = setInboxView(m)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 16})
	m = updated.(Model)
	m = sendTab(m)
	for i := 0; i < 35; i++ {
		m = sendRunes(m, 'j')
	}

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "ready") {
		t.Fatalf("footer should stay visible with long list, view=%q", view)
	}
	if !selectedLineContains(view, "task-36") {
		t.Fatalf("selected task should stay in visible window, view=%q", view)
	}
}

func TestLongProjectNavShouldKeepFooterAndSelectedVisible(t *testing.T) {
	projects := make([]string, 0, 50)
	for i := 1; i <= 50; i++ {
		projects = append(projects, fmt.Sprintf("proj-%02d", i))
	}
	r := &fakeTaskRepo{projects: projects}
	m := NewModelWithRepo(r)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 16})
	m = updated.(Model)
	for i := 0; i < 35; i++ {
		m = sendRunes(m, 'j')
	}
	row, ok := m.currentNavRow()
	if !ok {
		t.Fatalf("current nav row should exist")
	}
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "ready") {
		t.Fatalf("footer should stay visible with long nav, view=%q", view)
	}
	if !selectedLineContains(view, row.Label) {
		t.Fatalf("selected nav row should stay in visible window, want=%q view=%q", row.Label, view)
	}
}

func assertViewFits(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) > height {
		t.Fatalf("view line count = %d, want <= %d", len(lines), height)
	}

	maxWidth := 0
	for _, line := range lines {
		w := ansi.StringWidth(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	if maxWidth > width {
		t.Fatalf("view max line width = %d, want <= %d", maxWidth, width)
	}
}

type fakeTaskRepo struct {
	tasks    []domain.Task
	nextID   int64
	nowFunc  func() time.Time
	projects []string
}

func (f *fakeTaskRepo) Create(_ context.Context, task domain.Task) (int64, error) {
	if f.nextID <= 0 {
		f.nextID = 1
		for _, item := range f.tasks {
			if item.ID >= f.nextID {
				f.nextID = item.ID + 1
			}
		}
	}
	task.ID = f.nextID
	f.nextID++
	if task.Status == "" {
		task.Status = domain.StatusInbox
	}
	now := time.Now().UTC()
	if f.nowFunc != nil {
		now = f.nowFunc().UTC()
	}
	task.CreatedAt = now
	task.UpdatedAt = now
	f.tasks = append(f.tasks, task)
	return task.ID, nil
}

func (f *fakeTaskRepo) GetByID(_ context.Context, id int64) (domain.Task, error) {
	for _, task := range f.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return domain.Task{}, domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) List(_ context.Context, filter repo.TaskListFilter) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		if filter.Project != "" && task.Project != filter.Project {
			continue
		}
		out = append(out, task)
	}
	return out, nil
}

func (f *fakeTaskRepo) CreateProject(_ context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if !containsString(f.projects, name) {
		f.projects = append(f.projects, name)
		sort.Strings(f.projects)
	}
	return nil
}

func (f *fakeTaskRepo) ListProjects(context.Context) ([]string, error) {
	out := make([]string, 0, len(f.projects))
	out = append(out, f.projects...)
	sort.Strings(out)
	return out, nil
}

func (f *fakeTaskRepo) RenameProject(_ context.Context, oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	for i, name := range f.projects {
		if name == oldName {
			f.projects[i] = newName
		}
	}
	for i := range f.tasks {
		if f.tasks[i].Project == oldName {
			f.tasks[i].Project = newName
		}
	}
	sort.Strings(f.projects)
	return nil
}

func (f *fakeTaskRepo) DeleteProject(_ context.Context, name string) error {
	name = strings.TrimSpace(name)
	out := make([]string, 0, len(f.projects))
	for _, item := range f.projects {
		if item == name {
			continue
		}
		out = append(out, item)
	}
	f.projects = out
	for i := range f.tasks {
		if f.tasks[i].Project == name {
			f.tasks[i].Project = ""
		}
	}
	return nil
}

func (f *fakeTaskRepo) UpdateTitle(_ context.Context, id int64, title string) error {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks[i].Title = title
			return nil
		}
	}
	return domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) UpdateProject(_ context.Context, id int64, project string) error {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks[i].Project = project
			if project != "" && !containsString(f.projects, project) {
				f.projects = append(f.projects, project)
				sort.Strings(f.projects)
			}
			return nil
		}
	}
	return domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) UpdateDueAt(_ context.Context, id int64, dueAt *time.Time) error {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks[i].DueAt = dueAt
			return nil
		}
	}
	return domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) UpdatePriority(_ context.Context, id int64, priority string) error {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks[i].Priority = domain.NormalizePriority(priority)
			return nil
		}
	}
	return domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) SetStatus(_ context.Context, id int64, status domain.Status) error {
	for i := range f.tasks {
		if f.tasks[i].ID == id {
			f.tasks[i].Status = status
			if status == domain.StatusDone {
				now := time.Now().UTC()
				f.tasks[i].DoneAt = &now
			} else {
				f.tasks[i].DoneAt = nil
			}
			return nil
		}
	}
	return domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) MarkDone(_ context.Context, ids []int64) error {
	for _, id := range ids {
		for i := range f.tasks {
			if f.tasks[i].ID == id {
				f.tasks[i].Status = domain.StatusDone
				now := time.Now().UTC()
				f.tasks[i].DoneAt = &now
			}
		}
	}
	return nil
}

func (f *fakeTaskRepo) MarkDoing(_ context.Context, ids []int64) error {
	for _, id := range ids {
		for i := range f.tasks {
			if f.tasks[i].ID == id {
				f.tasks[i].Status = domain.StatusDoing
				f.tasks[i].DoneAt = nil
			}
		}
	}
	return nil
}

func (f *fakeTaskRepo) Reopen(_ context.Context, ids []int64) error {
	for _, id := range ids {
		for i := range f.tasks {
			if f.tasks[i].ID == id {
				f.tasks[i].Status = domain.StatusTodo
				f.tasks[i].DoneAt = nil
			}
		}
	}
	return nil
}

func (f *fakeTaskRepo) SoftDelete(_ context.Context, ids []int64) error {
	for _, id := range ids {
		for i := range f.tasks {
			if f.tasks[i].ID == id {
				f.tasks[i].Status = domain.StatusDeleted
			}
		}
	}
	return nil
}

func (f *fakeTaskRepo) Restore(_ context.Context, ids []int64) error {
	for _, id := range ids {
		for i := range f.tasks {
			if f.tasks[i].ID == id {
				f.tasks[i].Status = domain.StatusTodo
				f.tasks[i].DoneAt = nil
			}
		}
	}
	return nil
}

func (f *fakeTaskRepo) Purge(_ context.Context, ids []int64) error {
	drop := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		drop[id] = struct{}{}
	}
	out := make([]domain.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		if _, ok := drop[task.ID]; ok {
			continue
		}
		out = append(out, task)
	}
	f.tasks = out
	return nil
}

func setInboxView(m Model) Model {
	m.activeView = domain.ViewInbox
	m.reload()
	return m
}

func sendRunes(m Model, r ...rune) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: r})
	return updated.(Model)
}

func sendText(m Model, text string) Model {
	for _, r := range text {
		m = sendRunes(m, r)
	}
	return m
}

func sendEnter(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return updated.(Model)
}

func sendTab(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	return updated.(Model)
}

func sendBackspace(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	return updated.(Model)
}

func sendLeft(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	return updated.(Model)
}

func sendRight(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	return updated.(Model)
}

func sendCtrlB(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	return updated.(Model)
}

func sendCtrlF(m Model) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	return updated.(Model)
}

type fakeParseProvider struct {
	raw string
	err error
}

func (f fakeParseProvider) ParseTask(_ context.Context, _ string) (string, error) {
	return f.raw, f.err
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func selectedLineContains(view, target string) bool {
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, target) && strings.Contains(line, "> ") {
			return true
		}
	}
	return false
}

func lineContaining(view, target string) string {
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, target) {
			return line
		}
	}
	return ""
}
