package main

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "td",
	}
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
