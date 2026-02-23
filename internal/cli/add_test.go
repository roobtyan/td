package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"td/internal/config"
)

func TestAddAndList(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "add", "buy milk")
	if !strings.Contains(out, "created") {
		t.Fatalf("add output = %q, want contains %q", out, "created")
	}

	out = runCLI(t, cfg, "ls")
	if !strings.Contains(out, "buy milk") {
		t.Fatalf("ls output = %q, want contains %q", out, "buy milk")
	}
}

func runCLI(t *testing.T, cfg config.Config, args ...string) string {
	t.Helper()
	cmd := NewRootCmd(cfg)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.String()
}
