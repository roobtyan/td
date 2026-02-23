package usecase

import (
	"context"

	"td/internal/repo"
)

type ProjectUseCase struct {
	Repo repo.TaskRepository
}

func (u ProjectUseCase) List(ctx context.Context) ([]string, error) {
	return u.Repo.ListProjects(ctx)
}

func (u ProjectUseCase) Add(ctx context.Context, name string) error {
	return u.Repo.CreateProject(ctx, name)
}

func (u ProjectUseCase) Rename(ctx context.Context, oldName, newName string) error {
	return u.Repo.RenameProject(ctx, oldName, newName)
}

func (u ProjectUseCase) Delete(ctx context.Context, name string) error {
	return u.Repo.DeleteProject(ctx, name)
}
