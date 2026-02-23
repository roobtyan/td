package cli

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"td/internal/config"
)

func TestProjectCommandCRUD(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	_ = runCLI(t, cfg, "project", "add", "work")
	_ = runCLI(t, cfg, "project", "add", "home")

	out := runCLI(t, cfg, "project", "ls")
	if !strings.Contains(out, "home") || !strings.Contains(out, "work") {
		t.Fatalf("project ls output = %q, want contains home/work", out)
	}

	_ = runCLI(t, cfg, "project", "rename", "work", "office")
	out = runCLI(t, cfg, "project", "ls")
	if strings.Contains(out, "work") || !strings.Contains(out, "office") {
		t.Fatalf("project ls after rename = %q, want contains office and not work", out)
	}

	createOut := runCLI(t, cfg, "add", "prepare meeting", "-p", "office")
	matched := regexp.MustCompile(`created #(\d+)`).FindStringSubmatch(createOut)
	if len(matched) != 2 {
		t.Fatalf("cannot parse created id from output: %q", createOut)
	}
	id, err := strconv.ParseInt(matched[1], 10, 64)
	if err != nil {
		t.Fatalf("parse created id: %v", err)
	}

	_ = runCLI(t, cfg, "project", "rm", "office")

	repo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	task, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Project != "" {
		t.Fatalf("task project = %q, want empty after project rm", task.Project)
	}
}
