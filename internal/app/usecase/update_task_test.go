package usecase

import (
	"context"
	"testing"
	"time"

	"td/internal/domain"
	"td/internal/repo"
)

func TestUpdateTaskUseCaseMarkToday(t *testing.T) {
	stub := &updateTaskRepoStub{}
	uc := UpdateTaskUseCase{Repo: stub}

	ids := []int64{1, 2}
	if err := uc.MarkToday(context.Background(), ids); err != nil {
		t.Fatalf("mark today: %v", err)
	}
	if len(stub.markDoingIDs) != 2 || stub.markDoingIDs[0] != 1 || stub.markDoingIDs[1] != 2 {
		t.Fatalf("markDoingIDs = %v, want [1 2]", stub.markDoingIDs)
	}
}

func TestUpdateTaskUseCaseSetProject(t *testing.T) {
	stub := &updateTaskRepoStub{}
	uc := UpdateTaskUseCase{Repo: stub}

	if err := uc.SetProject(context.Background(), 7, "work"); err != nil {
		t.Fatalf("set project: %v", err)
	}
	if stub.projectID != 7 || stub.project != "work" {
		t.Fatalf("project update = (id=%d, project=%q), want (7, %q)", stub.projectID, stub.project, "work")
	}
}

func TestUpdateTaskUseCaseSetDueAt(t *testing.T) {
	stub := &updateTaskRepoStub{}
	uc := UpdateTaskUseCase{Repo: stub}

	due := time.Date(2026, 2, 24, 9, 30, 0, 0, time.UTC)
	if err := uc.SetDueAt(context.Background(), 9, &due); err != nil {
		t.Fatalf("set due: %v", err)
	}
	if stub.dueID != 9 {
		t.Fatalf("due id = %d, want 9", stub.dueID)
	}
	if stub.dueAt == nil || !stub.dueAt.Equal(due) {
		t.Fatalf("due = %v, want %v", stub.dueAt, due)
	}
}

func TestUpdateTaskUseCaseMarkProjectDone(t *testing.T) {
	stub := &updateTaskRepoStub{
		tasks: []domain.Task{
			{ID: 1, Status: domain.StatusInbox, Project: "work"},
			{ID: 2, Status: domain.StatusTodo, Project: "work"},
			{ID: 3, Status: domain.StatusDoing, Project: "work"},
			{ID: 4, Status: domain.StatusDone, Project: "work"},
			{ID: 5, Status: domain.StatusTodo, Project: "home"},
		},
	}
	uc := UpdateTaskUseCase{Repo: stub}

	done, err := uc.MarkProjectDone(context.Background(), "work")
	if err != nil {
		t.Fatalf("mark project done: %v", err)
	}
	if done != 3 {
		t.Fatalf("done count = %d, want 3", done)
	}
	if len(stub.markDoneIDs) != 3 || stub.markDoneIDs[0] != 1 || stub.markDoneIDs[1] != 2 || stub.markDoneIDs[2] != 3 {
		t.Fatalf("markDoneIDs = %v, want [1 2 3]", stub.markDoneIDs)
	}
}

type updateTaskRepoStub struct {
	projectID    int64
	project      string
	dueID        int64
	dueAt        *time.Time
	markDoingIDs []int64
	markDoneIDs  []int64
	tasks        []domain.Task
}

func (s *updateTaskRepoStub) Create(context.Context, domain.Task) (int64, error) {
	return 0, nil
}

func (s *updateTaskRepoStub) GetByID(context.Context, int64) (domain.Task, error) {
	return domain.Task{}, nil
}

func (s *updateTaskRepoStub) List(_ context.Context, filter repo.TaskListFilter) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if filter.Project != "" && task.Project != filter.Project {
			continue
		}
		out = append(out, task)
	}
	return out, nil
}

func (s *updateTaskRepoStub) CreateProject(context.Context, string) error {
	return nil
}

func (s *updateTaskRepoStub) ListProjects(context.Context) ([]string, error) {
	return nil, nil
}

func (s *updateTaskRepoStub) RenameProject(context.Context, string, string) error {
	return nil
}

func (s *updateTaskRepoStub) DeleteProject(context.Context, string) error {
	return nil
}

func (s *updateTaskRepoStub) UpdateTitle(context.Context, int64, string) error {
	return nil
}

func (s *updateTaskRepoStub) UpdateProject(_ context.Context, id int64, project string) error {
	s.projectID = id
	s.project = project
	return nil
}

func (s *updateTaskRepoStub) UpdateDueAt(_ context.Context, id int64, dueAt *time.Time) error {
	s.dueID = id
	s.dueAt = dueAt
	return nil
}

func (s *updateTaskRepoStub) MarkDone(_ context.Context, ids []int64) error {
	s.markDoneIDs = append([]int64(nil), ids...)
	return nil
}

func (s *updateTaskRepoStub) MarkDoing(_ context.Context, ids []int64) error {
	s.markDoingIDs = append([]int64(nil), ids...)
	return nil
}

func (s *updateTaskRepoStub) Reopen(context.Context, []int64) error {
	return nil
}

func (s *updateTaskRepoStub) SoftDelete(context.Context, []int64) error {
	return nil
}

func (s *updateTaskRepoStub) Restore(context.Context, []int64) error {
	return nil
}

func (s *updateTaskRepoStub) Purge(context.Context, []int64) error {
	return nil
}
