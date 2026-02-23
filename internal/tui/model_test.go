package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"td/internal/domain"
	"td/internal/repo"
)

func TestLayoutRatio(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "Inbox") {
		t.Fatalf("view should contain Inbox, got: %q", view)
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
	if !strings.Contains(view, "q quit") {
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

type fakeTaskRepo struct {
	tasks []domain.Task
}

func (f *fakeTaskRepo) Create(context.Context, domain.Task) (int64, error) {
	return 0, nil
}

func (f *fakeTaskRepo) GetByID(_ context.Context, id int64) (domain.Task, error) {
	for _, task := range f.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return domain.Task{}, domain.ErrTaskNotFound
}

func (f *fakeTaskRepo) List(_ context.Context, _ repo.TaskListFilter) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(f.tasks))
	out = append(out, f.tasks...)
	return out, nil
}

func (f *fakeTaskRepo) UpdateTitle(context.Context, int64, string) error { return nil }
func (f *fakeTaskRepo) MarkDone(context.Context, []int64) error          { return nil }
func (f *fakeTaskRepo) Reopen(context.Context, []int64) error            { return nil }
func (f *fakeTaskRepo) SoftDelete(context.Context, []int64) error        { return nil }
func (f *fakeTaskRepo) Restore(context.Context, []int64) error           { return nil }
func (f *fakeTaskRepo) Purge(context.Context, []int64) error             { return nil }
