package cli

import (
	"github.com/spf13/cobra"

	"td/internal/buildinfo"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("version: %s\n", buildinfo.Version)
			cmd.Printf("commit: %s\n", buildinfo.Commit)
			cmd.Printf("built: %s\n", buildinfo.Date)
		},
	}
	return cmd
}
