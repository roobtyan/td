package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newAddCmd(cfg config.Config) *cobra.Command {
	var (
		project  string
		priority string
	)

	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Add a task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.AddTaskUseCase{Repo: repo}
			task, err := uc.Execute(cmd.Context(), usecase.AddTaskInput{
				Title:    strings.Join(args, " "),
				Project:  project,
				Priority: priority,
			})
			if err != nil {
				return err
			}

			cmd.Printf("created #%d %s\n", task.ID, task.Title)
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project")
	cmd.Flags().StringVarP(&priority, "priority", "P", "P2", "priority")
	return cmd
}
