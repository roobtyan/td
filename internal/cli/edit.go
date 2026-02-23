package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newEditCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <id> <title>",
		Short: "Edit task title",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args[:1])
			if err != nil {
				return err
			}
			title := strings.Join(args[1:], " ")

			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.UpdateTaskUseCase{Repo: repo}
			if err := uc.EditTitle(cmd.Context(), ids[0], title); err != nil {
				return err
			}
			cmd.Printf("edited #%d\n", ids[0])
			return nil
		},
	}
	return cmd
}
