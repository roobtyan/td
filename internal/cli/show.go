package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"td/internal/config"
)

func newShowCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			task, err := repo.GetByID(cmd.Context(), id)
			if err != nil {
				return err
			}
			cmd.Printf("%d\n", task.ID)
			cmd.Printf("title: %s\n", task.Title)
			cmd.Printf("status: %s\n", task.Status)
			cmd.Printf("project: %s\n", task.Project)
			cmd.Printf("priority: %s\n", task.Priority)
			if task.Notes != "" {
				cmd.Printf("notes: %s\n", task.Notes)
			}
			return nil
		},
	}
	return cmd
}
