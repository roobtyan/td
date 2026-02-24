package cli

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newAddCmd(cfg config.Config) *cobra.Command {
	var (
		project  string
		priority string
		fromClip bool
		useAI    bool
		dueRaw   string
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

			var dueAt *time.Time
			if strings.TrimSpace(dueRaw) != "" {
				parsed, err := parseDueInput(dueRaw, time.Local)
				if err != nil {
					return err
				}
				dueAt = &parsed
			}

			var taskTitle string
			var taskID int64
			if fromClip {
				var aiParser *usecase.AIParseTaskUseCase
				if useAI {
					aiParser = newAIParseTaskUseCase(cfg)
				}
				uc := usecase.AddFromClipboardUseCase{
					Repo:     repo,
					AIParser: aiParser,
					Project:  project,
					Priority: priority,
					DueAt:    dueAt,
				}
				clipText := strings.Join(args, " ")
				task, err := uc.AddFromClipboard(cmd.Context(), clipText, useAI)
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
					DueAt:    dueAt,
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
	cmd.Flags().StringVar(&dueRaw, "due", "", "due datetime, supports YYYY-MM-DD or YYYY-MM-DD HH:MM")
	cmd.Flags().BoolVar(&fromClip, "clip", false, "create from clipboard")
	cmd.Flags().BoolVar(&useAI, "ai", false, "parse clipboard with AI and fallback to rules")
	return cmd
}
