package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"td/internal/buildinfo"
	"td/internal/config"
)

func TestVersionCommandShowsBuildInfo(t *testing.T) {
	oldVersion := buildinfo.Version
	oldCommit := buildinfo.Commit
	oldDate := buildinfo.Date
	t.Cleanup(func() {
		buildinfo.Version = oldVersion
		buildinfo.Commit = oldCommit
		buildinfo.Date = oldDate
	})

	buildinfo.Version = "v1.2.3"
	buildinfo.Commit = "abc1234"
	buildinfo.Date = "2026-02-23"

	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "version")
	if !strings.Contains(out, "v1.2.3") {
		t.Fatalf("version output = %q, want contains %q", out, "v1.2.3")
	}
	if !strings.Contains(out, "abc1234") {
		t.Fatalf("version output = %q, want contains %q", out, "abc1234")
	}
}

func TestUpgradeCommandExists(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "upgrade", "--help")
	if !strings.Contains(out, "upgrade") {
		t.Fatalf("upgrade help output = %q, want contains %q", out, "upgrade")
	}
}
