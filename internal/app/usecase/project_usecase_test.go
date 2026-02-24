package usecase

import (
	"context"
	"testing"
	"time"

	"td/internal/domain"
	"td/internal/repo"
)

func TestProjectUseCaseCRUD(t *testing.T) {
	stub := &projectRepoStub{
		projects: []string{"home", "work"},
	}
	uc := ProjectUseCase{Repo: stub}

	list, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(list) != 2 || list[0] != "home" || list[1] != "work" {
		t.Fatalf("projects = %v, want [home work]", list)
	}

	if err := uc.Add(context.Background(), "study"); err != nil {
		t.Fatalf("add project: %v", err)
	}
	if stub.created != "study" {
		t.Fatalf("created = %q, want %q", stub.created, "study")
	}

	if err := uc.Rename(context.Background(), "work", "office"); err != nil {
		t.Fatalf("rename project: %v", err)
	}
	if stub.renamedOld != "work" || stub.renamedNew != "office" {
		t.Fatalf("rename = (%q, %q), want (%q, %q)", stub.renamedOld, stub.renamedNew, "work", "office")
	}

	if err := uc.Delete(context.Background(), "office"); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if stub.deleted != "office" {
		t.Fatalf("deleted = %q, want %q", stub.deleted, "office")
	}
}

type projectRepoStub struct {
	projects   []string
	created    string
	renamedOld string
	renamedNew string
	deleted    string
}

func (s *projectRepoStub) Create(context.Context, domain.Task) (int64, error) {
	return 0, nil
}

func (s *projectRepoStub) GetByID(context.Context, int64) (domain.Task, error) {
	return domain.Task{}, nil
}

func (s *projectRepoStub) List(context.Context, repo.TaskListFilter) ([]domain.Task, error) {
	return nil, nil
}

func (s *projectRepoStub) CreateProject(_ context.Context, name string) error {
	s.created = name
	return nil
}

func (s *projectRepoStub) ListProjects(context.Context) ([]string, error) {
	out := make([]string, 0, len(s.projects))
	out = append(out, s.projects...)
	return out, nil
}

func (s *projectRepoStub) RenameProject(_ context.Context, oldName, newName string) error {
	s.renamedOld = oldName
	s.renamedNew = newName
	return nil
}

func (s *projectRepoStub) DeleteProject(_ context.Context, name string) error {
	s.deleted = name
	return nil
}

func (s *projectRepoStub) UpdateTitle(context.Context, int64, string) error { return nil }
func (s *projectRepoStub) UpdateProject(context.Context, int64, string) error {
	return nil
}
func (s *projectRepoStub) UpdateDueAt(context.Context, int64, *time.Time) error {
	return nil
}
func (s *projectRepoStub) UpdatePriority(context.Context, int64, string) error   { return nil }
func (s *projectRepoStub) SetStatus(context.Context, int64, domain.Status) error { return nil }
func (s *projectRepoStub) MarkDone(context.Context, []int64) error               { return nil }
func (s *projectRepoStub) MarkDoing(context.Context, []int64) error              { return nil }
func (s *projectRepoStub) Reopen(context.Context, []int64) error                 { return nil }
func (s *projectRepoStub) SoftDelete(context.Context, []int64) error             { return nil }
func (s *projectRepoStub) Restore(context.Context, []int64) error                { return nil }
func (s *projectRepoStub) Purge(context.Context, []int64) error                  { return nil }
