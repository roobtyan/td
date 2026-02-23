package cli

import (
	"time"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newDueCmd(cfg config.Config) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "due <id> <datetime>",
		Short: "Set task due datetime",
		Args: func(cmd *cobra.Command, args []string) error {
			if clear {
				return cobra.ExactArgs(1)(cmd, args)
			}
			return cobra.ExactArgs(2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args[:1])
			if err != nil {
				return err
			}
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			var dueAt *time.Time
			if !clear {
				parsed, err := parseDueInput(args[1], time.Local)
				if err != nil {
					return err
				}
				dueAt = &parsed
			}

			uc := usecase.UpdateTaskUseCase{Repo: repo}
			if err := uc.SetDueAt(cmd.Context(), ids[0], dueAt); err != nil {
				return err
			}
			if clear {
				cmd.Printf("cleared due #%d\n", ids[0])
			} else {
				cmd.Printf("set due #%d\n", ids[0])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "clear due datetime")
	return cmd
}
