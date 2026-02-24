package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
	"td/internal/domain"
)

func newPriorityCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "priority <id> <P1|P2|P3|P4>",
		Aliases: []string{"pri"},
		Short:   "Set task priority",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args[:1])
			if err != nil {
				return err
			}
			priority := domain.NormalizePriority(strings.TrimSpace(args[1]))
			if !domain.IsValidPriority(priority) {
				return domain.ErrInvalidPriority
			}

			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.UpdateTaskUseCase{Repo: repo}
			if err := uc.SetPriority(cmd.Context(), ids[0], priority); err != nil {
				return err
			}
			cmd.Printf("priority #%d -> %s\n", ids[0], priority)
			return nil
		},
	}
	return cmd
}
