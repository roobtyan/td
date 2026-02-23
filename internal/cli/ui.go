package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"td/internal/config"
	"td/internal/tui"
)

func newUICmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Open terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, closer, err := openTaskRepo(cfg)
			if err != nil {
				return err
			}
			defer closeDB(closer)

			program := tea.NewProgram(tui.NewModelWithRepo(repo))
			_, err = program.Run()
			return err
		},
	}
	return cmd
}
