package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"td/internal/app/usecase"
	"td/internal/config"
)

func newLsCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			uc := usecase.ListTaskUseCase{Repo: repo}
			tasks, err := uc.Execute(cmd.Context(), usecase.ListTaskInput{})
			if err != nil {
				return err
			}
			for _, task := range tasks {
				cmd.Println(formatTaskLine(task.ID, string(task.Status), task.Title))
			}
			return nil
		},
	}
	return cmd
}

func formatTaskLine(id int64, status, title string) string {
	return fmt.Sprintf("%d\t[%s]\t%s", id, status, title)
}
