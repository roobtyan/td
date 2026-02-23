package cli

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"td/internal/config"
)

func TestAddWithDueFlag(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	out := runCLI(t, cfg, "add", "prepare slides", "--due", "2026-02-24 09:30")
	matched := regexp.MustCompile(`created #(\d+)`).FindStringSubmatch(out)
	if len(matched) != 2 {
		t.Fatalf("cannot parse created id from output: %q", out)
	}
	id, err := strconv.ParseInt(matched[1], 10, 64)
	if err != nil {
		t.Fatalf("parse created id: %v", err)
	}

	repo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	task, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.DueAt == nil {
		t.Fatalf("due_at should not be nil")
	}
}

func TestTodayAndDueCommand(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	id := createViaCLI(t, cfg, "task for today")
	idStr := strconv.FormatInt(id, 10)

	_ = runCLI(t, cfg, "today", idStr)
	_ = runCLI(t, cfg, "due", idStr, "2026-02-25 18:00")

	repo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	task, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != "doing" {
		t.Fatalf("status = %s, want doing", task.Status)
	}
	if task.DueAt == nil {
		t.Fatalf("due_at should not be nil")
	}
	want := time.Date(2026, 2, 25, 18, 0, 0, 0, time.Local).UTC()
	if !task.DueAt.Equal(want) {
		t.Fatalf("due_at = %s, want %s", task.DueAt.UTC(), want)
	}
}

func TestDueCommandShouldSupportCompactDateTime(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	id := createViaCLI(t, cfg, "task compact due")
	idStr := strconv.FormatInt(id, 10)
	_ = runCLI(t, cfg, "due", idStr, "202602051122")

	repo, closer, err := openTaskRepo(cfg)
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer closeDB(closer)

	task, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.DueAt == nil {
		t.Fatalf("due_at should not be nil")
	}
	want := time.Date(2026, 2, 5, 11, 22, 0, 0, time.Local).UTC()
	if !task.DueAt.Equal(want) {
		t.Fatalf("due_at = %s, want %s", task.DueAt.UTC(), want)
	}
}
