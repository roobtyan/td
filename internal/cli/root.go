package cli

import (
	"database/sql"
	"os"

	"github.com/spf13/cobra"

	"td/internal/config"
	"td/internal/repo/sqlite"
)

func NewRootCmd(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "td",
	}
	cmd.AddCommand(newAddCmd(cfg))
	cmd.AddCommand(newProjectCmd(cfg))
	cmd.AddCommand(newLsCmd(cfg))
	cmd.AddCommand(newShowCmd(cfg))
	cmd.AddCommand(newDoneCmd(cfg))
	cmd.AddCommand(newReopenCmd(cfg))
	cmd.AddCommand(newEditCmd(cfg))
	cmd.AddCommand(newTodayCmd(cfg))
	cmd.AddCommand(newDueCmd(cfg))
	cmd.AddCommand(newPriorityCmd(cfg))
	cmd.AddCommand(newRmCmd(cfg))
	cmd.AddCommand(newRestoreCmd(cfg))
	cmd.AddCommand(newPurgeCmd(cfg))
	cmd.AddCommand(newUICmd(cfg))
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newUpgradeCmd(cfg))
	cmd.AddCommand(newConfigCmd(cfg))
	return cmd
}

func openTaskRepo(cfg config.Config) (*sqlite.TaskRepository, func() error, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, nil, err
	}
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, nil, err
	}
	return sqlite.NewTaskRepository(db), func() error {
		return db.Close()
	}, nil
}

func closeDB(closer func() error) error {
	if closer == nil {
		return nil
	}
	if err := closer(); err != nil && err != sql.ErrConnDone {
		return err
	}
	return nil
}
