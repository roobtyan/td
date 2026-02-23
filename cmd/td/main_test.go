package main

import "testing"

func TestMainHelp(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}
}
