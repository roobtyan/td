package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"td/internal/config"
)

func TestPriorityCommandShouldUpdateTaskPriority(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	id := createViaCLIWithArgs(t, cfg, "task-priority", "-P", "P4")
	idStr := strconv.FormatInt(id, 10)
	_ = runCLI(t, cfg, "priority", idStr, "P1")

	repo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	task, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Priority != "P1" {
		t.Fatalf("task priority = %q, want %q", task.Priority, "P1")
	}
}

func TestPriorityCommandShouldValidateInput(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	id := createViaCLI(t, cfg, "task-priority-invalid")
	idStr := strconv.FormatInt(id, 10)

	out, err := runCLIWithErr(cfg, "priority", idStr, "PX")
	if err == nil {
		t.Fatalf("priority command should fail for invalid input")
	}
	if !strings.Contains(out, "invalid priority") {
		t.Fatalf("output should contain invalid priority, got %q", out)
	}
}

func runCLIWithErr(cfg config.Config, args ...string) (string, error) {
	cmd := NewRootCmd(cfg)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}
