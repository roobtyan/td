package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"td/internal/buildinfo"
	"td/internal/config"
	"td/internal/updater"
)

func newUpgradeCmd(cfg config.Config) *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Check GitHub release and self-upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.New(updater.Options{
				Owner:          buildinfo.RepoOwner,
				Repo:           buildinfo.RepoName,
				CurrentVersion: buildinfo.Version,
				GitHubToken:    resolveGitHubToken(cfg),
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

func resolveGitHubToken(cfg config.Config) string {
	if token := strings.TrimSpace(os.Getenv("GH_TOKEN")); token != "" {
		return token
	}
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		return token
	}
	userCfg, err := config.LoadUserConfig(cfg.ConfigToml)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(userCfg.GitHub.Token)
}
