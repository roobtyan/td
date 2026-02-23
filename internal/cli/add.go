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
		fromClip bool
	)

	cmd := &cobra.Command{
		Use:   "add <text>",
		Short: "Add a task",
		Args: func(cmd *cobra.Command, args []string) error {
			if fromClip {
				return nil
			}
			return cobra.MinimumNArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			var taskTitle string
			var taskID int64
			if fromClip {
				uc := usecase.AddFromClipboardUseCase{
					Repo:     repo,
					Project:  project,
					Priority: priority,
				}
				clipText := strings.Join(args, " ")
				task, err := uc.AddFromClipboard(cmd.Context(), clipText, false)
				if err != nil {
					return err
				}
				taskID = task.ID
				taskTitle = task.Title
			} else {
				uc := usecase.AddTaskUseCase{Repo: repo}
				task, err := uc.Execute(cmd.Context(), usecase.AddTaskInput{
					Title:    strings.Join(args, " "),
					Project:  project,
					Priority: priority,
				})
				if err != nil {
					return err
				}
				taskID = task.ID
				taskTitle = task.Title
			}

			cmd.Printf("created #%d %s\n", taskID, taskTitle)
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project")
	cmd.Flags().StringVarP(&priority, "priority", "P", "P2", "priority")
	cmd.Flags().BoolVar(&fromClip, "clip", false, "create from clipboard")
	return cmd
}
