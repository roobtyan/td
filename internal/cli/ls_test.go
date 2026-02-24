package cli

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"td/internal/config"
)

func TestLsDefaultShouldHideDeletedAndShowProjectDuePriorityAndSort(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	idDoing := createViaCLIWithArgs(t, cfg, "alpha-doing", "-p", "alpha", "-P", "P1")
	idTodo := createViaCLIWithArgs(t, cfg, "alpha-todo", "-p", "alpha", "-P", "P2")
	idDone := createViaCLIWithArgs(t, cfg, "alpha-done", "-p", "alpha", "-P", "P3")
	idBetaTodo := createViaCLIWithArgs(t, cfg, "beta-todo", "-p", "beta", "-P", "P4")
	idDeleted := createViaCLIWithArgs(t, cfg, "beta-deleted", "-p", "beta", "-P", "P2")

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
		formatTaskLine(idDoing, "doing", "alpha-doing", "alpha", mustParseLocalTime(t, "2026-02-24 08:00"), "P1"),
		formatTaskLine(idTodo, "todo", "alpha-todo", "alpha", mustParseLocalTime(t, "2026-02-25 09:30"), "P2"),
		formatTaskLine(idDone, "done", "alpha-done", "alpha", nil, "P3"),
		formatTaskLine(idBetaTodo, "todo", "beta-todo", "beta", nil, "P4"),
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d mismatch, got %q want %q", i, lines[i], want[i])
		}
		if strings.Contains(lines[i], "\t") {
			t.Fatalf("line %d should not contain tab, got %q", i, lines[i])
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

	if !strings.Contains(out, "work-overdue") || !strings.Contains(out, "work-today") {
		t.Fatalf("ls today should include overdue/today todo, got %q", out)
	}
}

func TestLsTodayShouldSortByPriorityThenDue(t *testing.T) {
	tdHome := t.TempDir()
	cfg := config.Default()
	cfg.HomeDir = tdHome
	cfg.DataDir = filepath.Join(tdHome, "data")
	cfg.DBPath = filepath.Join(cfg.DataDir, "td.db")

	now := time.Now().Local()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due0900 := dayStart.Add(9 * time.Hour).Format("2006-01-02 15:04")
	due1100 := dayStart.Add(11 * time.Hour).Format("2006-01-02 15:04")
	due1300 := dayStart.Add(13 * time.Hour).Format("2006-01-02 15:04")

	idP3 := createViaCLIWithArgs(t, cfg, "task-p3", "-p", "work", "-P", "P3")
	idP1 := createViaCLIWithArgs(t, cfg, "task-p1", "-p", "work", "-P", "P1")
	idP2 := createViaCLIWithArgs(t, cfg, "task-p2", "-p", "work", "-P", "P2")

	_ = runCLI(t, cfg, "due", strconv.FormatInt(idP3, 10), due1100)
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idP1, 10), due1300)
	_ = runCLI(t, cfg, "due", strconv.FormatInt(idP2, 10), due0900)

	out := runCLI(t, cfg, "ls", "today")
	lines := nonEmptyLines(out)
	if len(lines) != 3 {
		t.Fatalf("ls today line count = %d, want 3, output=%q", len(lines), out)
	}

	if !strings.Contains(lines[0], "task-p1") {
		t.Fatalf("first line should be highest priority P1 task, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "task-p2") {
		t.Fatalf("second line should be P2 task, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "task-p3") {
		t.Fatalf("third line should be P3 task, got %q", lines[2])
	}
}

func TestFormatTaskLineShouldAlignColumnsWithoutHeader(t *testing.T) {
	lineA := formatTaskLine(
		1,
		"todo",
		"short",
		"proj-a",
		mustParseLocalTime(t, "2026-02-24 08:00"),
		"P1",
	)
	lineB := formatTaskLine(
		2,
		"doing",
		"very-very-long-title",
		"proj-b",
		nil,
		"P4",
	)

	if strings.Contains(lineA, "\t") || strings.Contains(lineB, "\t") {
		t.Fatalf("formatted line should not contain tab, got %q / %q", lineA, lineB)
	}

	if ansi.StringWidth(lineA) != ansi.StringWidth(lineB) {
		t.Fatalf("line widths should align, got %d and %d", ansi.StringWidth(lineA), ansi.StringWidth(lineB))
	}

	idxProjectA := strings.Index(lineA, "proj-a")
	idxProjectB := strings.Index(lineB, "proj-b")
	if idxProjectA != idxProjectB {
		t.Fatalf("project column should align, idxA=%d idxB=%d, lineA=%q lineB=%q", idxProjectA, idxProjectB, lineA, lineB)
	}

	idxPriorityA := strings.LastIndex(lineA, "P1")
	idxPriorityB := strings.LastIndex(lineB, "P4")
	if idxPriorityA != idxPriorityB {
		t.Fatalf("priority column should align, idxA=%d idxB=%d, lineA=%q lineB=%q", idxPriorityA, idxPriorityB, lineA, lineB)
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
