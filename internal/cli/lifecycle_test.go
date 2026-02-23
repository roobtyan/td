package cli

import (
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"td/internal/config"
)

func TestLifecycleCommands(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	id := createViaCLI(t, cfg, "task A")
	idStr := strconv.FormatInt(id, 10)

	_ = runCLI(t, cfg, "done", idStr)
	_ = runCLI(t, cfg, "reopen", idStr)
	_ = runCLI(t, cfg, "rm", idStr)
	_ = runCLI(t, cfg, "restore", idStr)
	_ = runCLI(t, cfg, "rm", idStr)
	_ = runCLI(t, cfg, "purge", idStr)
}

func createViaCLI(t *testing.T, cfg config.Config, title string) int64 {
	t.Helper()
	out := runCLI(t, cfg, "add", title)
	matched := regexp.MustCompile(`created #(\d+)`).FindStringSubmatch(out)
	if len(matched) != 2 {
		t.Fatalf("cannot parse created id from output: %q", out)
	}
	id, err := strconv.ParseInt(matched[1], 10, 64)
	if err != nil {
		t.Fatalf("parse created id: %v", err)
	}
	return id
}
