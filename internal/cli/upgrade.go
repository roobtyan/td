package cli

import (
	"github.com/spf13/cobra"

	"td/internal/buildinfo"
	"td/internal/updater"
)

func newUpgradeCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Check GitHub release and self-upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.New(updater.Options{
				Owner:          buildinfo.RepoOwner,
				Repo:           buildinfo.RepoName,
				CurrentVersion: buildinfo.Version,
			})

			if checkOnly {
				result, err := u.Check(cmd.Context())
				if err != nil {
					return err
				}
				if result.HasUpdate {
					cmd.Printf("update available: %s -> %s\n", result.CurrentVersion, result.LatestVersion)
					cmd.Printf("run `td upgrade` to apply\n")
					return nil
				}
				cmd.Printf("already latest: %s\n", result.CurrentVersion)
				return nil
			}

			result, err := u.Upgrade(cmd.Context())
			if err != nil {
				return err
			}
			if result.Updated {
				cmd.Printf("upgraded: %s -> %s\n", result.CurrentVersion, result.LatestVersion)
				return nil
			}

			cmd.Printf("already latest: %s\n", result.CurrentVersion)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "check for updates only")
	return cmd
}
