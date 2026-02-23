package cli

import (
	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newProjectCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}
	cmd.AddCommand(newProjectLsCmd(cfg))
	cmd.AddCommand(newProjectAddCmd(cfg))
	cmd.AddCommand(newProjectRenameCmd(cfg))
	cmd.AddCommand(newProjectRmCmd(cfg))
	return cmd
}

func newProjectLsCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.ProjectUseCase{Repo: repo}
			projects, err := uc.List(cmd.Context())
			if err != nil {
				return err
			}
			for _, name := range projects {
				cmd.Println(name)
			}
			return nil
		},
	}
}

func newProjectAddCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Create project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.ProjectUseCase{Repo: repo}
			if err := uc.Add(cmd.Context(), args[0]); err != nil {
				return err
			}
			cmd.Printf("created project %s\n", args[0])
			return nil
		},
	}
}

func newProjectRenameCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.ProjectUseCase{Repo: repo}
			if err := uc.Rename(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			cmd.Printf("renamed project %s -> %s\n", args[0], args[1])
			return nil
		},
	}
}

func newProjectRmCmd(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete project and detach its tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.ProjectUseCase{Repo: repo}
			if err := uc.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			cmd.Printf("deleted project %s\n", args[0])
			return nil
		},
	}
}
