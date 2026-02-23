package main

import (
	"os"

	"github.com/spf13/cobra"

	"td/internal/cli"
	"td/internal/config"
)

func NewRootCmd() *cobra.Command {
	return cli.NewRootCmd(config.Default())
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
