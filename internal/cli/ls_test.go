package cli

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"td/internal/config"
)

func TestLsDefaultShouldHideDeletedAndShowProjectDueAndSort(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	idDoing := createViaCLIWithArgs(t, cfg, "alpha-doing", "-p", "alpha")
	idTodo := createViaCLIWithArgs(t, cfg, "alpha-todo", "-p", "alpha")
	idDone := createViaCLIWithArgs(t, cfg, "alpha-done", "-p", "alpha")
	idBetaTodo := createViaCLIWithArgs(t, cfg, "beta-todo", "-p", "beta")
	idDeleted := createViaCLIWithArgs(t, cfg, "beta-deleted", "-p", "beta")

	_ = runCLI(t, cfg, "today", strconv.FormatInt(idDoing, 10))
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idDoing, 10), "2026-02-24 08:00")
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idTodo, 10), "2026-02-25 09:30")
	_ = runCLI(t, cfg, "done", strconv.FormatInt(idDone, 10))
	_ = runCLI(t, cfg, "rm", strconv.FormatInt(idDeleted, 10))

	out := runCLI(t, cfg, "ls")
	if strings.Contains(out, "beta-deleted") {
		t.Fatalf("ls output should hide deleted task, got %q", out)
	}

	lines := nonEmptyLines(out)
	if len(lines) != 4 {
		t.Fatalf("ls line count = %d, want 4, output=%q", len(lines), out)
	}

	want := []string{
		formatTaskLine(idDoing, "doing", "alpha-doing", "alpha", mustParseLocalTime(t, "2026-02-24 08:00")),
		formatTaskLine(idTodo, "todo", "alpha-todo", "alpha", mustParseLocalTime(t, "2026-02-25 09:30")),
		formatTaskLine(idDone, "done", "alpha-done", "alpha", nil),
		formatTaskLine(idBetaTodo, "todo", "beta-todo", "beta", nil),
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d mismatch, got %q want %q", i, lines[i], want[i])
		}
	}
}

func TestLsTodayShouldOnlyShowTodayTasks(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	now := time.Now().Local()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	overdue := dayStart.Add(-2 * time.Hour).Format("2006-01-02 15:04")
	today := dayStart.Add(12 * time.Hour).Format("2006-01-02 15:04")
	future := dayStart.Add(36 * time.Hour).Format("2006-01-02 15:04")

	idDoing := createViaCLIWithArgs(t, cfg, "work-doing", "-p", "work")
	idOverdue := createViaCLIWithArgs(t, cfg, "work-overdue", "-p", "work")
	idToday := createViaCLIWithArgs(t, cfg, "work-today", "-p", "work")
	idFuture := createViaCLIWithArgs(t, cfg, "work-future", "-p", "work")
	idInbox := createViaCLI(t, cfg, "inbox-today")

	_ = runCLI(t, cfg, "today", strconv.FormatInt(idDoing, 10))
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idOverdue, 10), overdue)
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idToday, 10), today)
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idFuture, 10), future)
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idInbox, 10), today)

	out := runCLI(t, cfg, "ls", "today")
	if strings.Contains(out, "work-future") {
		t.Fatalf("ls today output should hide future task, got %q", out)
	}
	if strings.Contains(out, "inbox-today") {
		t.Fatalf("ls today output should hide inbox task, got %q", out)
	}

	lines := nonEmptyLines(out)
	if len(lines) != 3 {
		t.Fatalf("ls today line count = %d, want 3, output=%q", len(lines), out)
	}

	if !strings.Contains(lines[0], "[doing]\twork-doing") {
		t.Fatalf("first line should be doing task, got %q", lines[0])
	}
	if !strings.Contains(out, "work-overdue") || !strings.Contains(out, "work-today") {
		t.Fatalf("ls today should include overdue/today todo, got %q", out)
	}
}

func createViaCLIWithArgs(t *testing.T, cfg config.Config, title string, args ...string) int64 {
	t.Helper()
	argv := []string{"add", title}
	argv = append(argv, args...)
	out := runCLI(t, cfg, argv...)
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

func nonEmptyLines(out string) []string {
	raw := strings.Split(strings.TrimSpace(out), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func mustParseLocalTime(t *testing.T, raw string) *time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04", raw, time.Local)
	if err != nil {
		t.Fatalf("parse local time %q: %v", raw, err)
	}
	return &v
}
