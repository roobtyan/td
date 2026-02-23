package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"td/internal/config"
	taskrepo "td/internal/repo"
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

func TestAddWithProjectShouldCreateTodo(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "add", "task in project", "-p", "work")
	if !strings.Contains(out, "created") {
		t.Fatalf("add output = %q, want contains created", out)
	}

	taskRepo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	tasks, err := taskRepo.List(context.Background(), taskrepo.TaskListFilter{})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(tasks))
	}
	if tasks[0].Status != "todo" {
		t.Fatalf("status = %s, want todo", tasks[0].Status)
	}
	if tasks[0].Project != "work" {
		t.Fatalf("project = %q, want %q", tasks[0].Project, "work")
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
