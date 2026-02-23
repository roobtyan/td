package cli

import (
	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newTodayCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "today <id...>",
		Short: "Set tasks to today by marking them doing",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.UpdateTaskUseCase{Repo: repo}
			if err := uc.MarkToday(cmd.Context(), ids); err != nil {
				return err
			}
			cmd.Printf("today %d task(s)\n", len(ids))
			return nil
		},
	}
	return cmd
}
