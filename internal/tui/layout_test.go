package tui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"td/internal/domain"
)

func TestPadRightWithANSIShouldKeepDisplayWidth(t *testing.T) {
	styled := "\x1b[31m1234567890\x1b[0m"

	got := padRight(styled, 8)
	if ansi.StringWidth(got) != 8 {
		t.Fatalf("padRight display width = %d, want 8, raw=%q", ansi.StringWidth(got), got)
	}
	if !strings.Contains(got, "...") {
		t.Fatalf("padRight should include ellipsis for truncation, got %q", got)
	}
	if ansi.Strip(got) != "12345..." {
		t.Fatalf("padRight strip = %q, want %q, raw=%q", ansi.Strip(got), "12345...", got)
	}
}

func TestProgressPercentNoTasksShouldBeHundred(t *testing.T) {
	if got := progressPercent(0, 0); got != 100 {
		t.Fatalf("progressPercent(0,0) = %d, want 100", got)
	}
}

func TestHeaderLogoAndInfoShouldUseSameForegroundTone(t *testing.T) {
	if logoTextColor != headerTextColor {
		t.Fatalf("logo/info foreground mismatch: logo=%v info=%v", logoTextColor, headerTextColor)
	}
}

func TestHeaderShouldNotContainMidResetBetweenSegments(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	header := renderHeader(
		0,
		0,
		domain.ViewInbox,
		0,
		0,
		0,
		0,
		0,
		time.Date(2026, 2, 23, 18, 24, 0, 0, time.Local),
		130,
	)

	if strings.Contains(header, "\x1b[0m  \x1b[1;") {
		t.Fatalf("header contains mid-reset artifacts: %q", header)
	}
}

func TestFooterShouldNotContainMidResetBetweenSegments(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	footer := renderFooter("ready", focusNav, 130, domain.ViewInbox)
	if strings.Contains(footer, "\x1b[0m  \x1b[38;") {
		t.Fatalf("footer contains mid-reset artifacts: %q", footer)
	}
}

func TestFooterShouldContainHighlightStyle(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	footer := renderFooter("ready", focusNav, 130, domain.ViewInbox)
	if !containsANSIColor(footer, "52;211;153") {
		t.Fatalf("footer should contain ready highlight color, footer=%q", footer)
	}
}

func TestFooterShouldRestoreMutedBaseColorAfterTokenHighlight(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	footer := renderFooter("ready", focusNav, 130, domain.ViewInbox)
	if !strings.Contains(footer, "38;2;147;161;176m") {
		t.Fatalf("footer should contain muted base color marker, footer=%q", footer)
	}
	if strings.Contains(footer, "\x1b[0m ready") {
		t.Fatalf("footer should not reset to terminal default before normal text, footer=%q", footer)
	}
}

func TestHeaderShouldContainHighlightStyle(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	header := renderHeader(
		0,
		0,
		domain.ViewInbox,
		0,
		0,
		0,
		0,
		0,
		time.Date(2026, 2, 23, 18, 30, 0, 0, time.Local),
		140,
	)
	if !containsHighlightedLabel(header, "Time:") {
		t.Fatalf("header should highlight Time label, header=%q", header)
	}
	if !containsHighlightedLabel(header, "View:") {
		t.Fatalf("header should highlight View label, header=%q", header)
	}
	if !containsHighlightedLabel(header, "Doing:") {
		t.Fatalf("header should highlight Doing label, header=%q", header)
	}
}

func TestListPaneShouldNotContainMidResetArtifacts(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	lines := renderList(nil, 0, false, 60, 12, domain.ViewInbox, time.Local)
	block := strings.Join(lines, "\n")
	if strings.Contains(block, "\x1b[1;38;2;217;226;236mTasks") {
		t.Fatalf("list title should not use nested line style render, block=%q", block)
	}
	if strings.Contains(block, "\x1b[38;2;226;232;240m[empty]") {
		t.Fatalf("list pane should avoid mid resets, block=%q", block)
	}
}

func TestRenderBoxShouldKeepFixedHeightWhenContentOverflows(t *testing.T) {
	content := strings.Repeat("this-line-is-too-long-for-pane-width\n", 24)
	block := renderBox(listBoxStyle, content, 36, 10)
	lines := strings.Split(block, "\n")
	if len(lines) != 10 {
		t.Fatalf("renderBox line count = %d, want 10", len(lines))
	}
	last := ansi.Strip(lines[len(lines)-1])
	if !strings.Contains(last, "┘") {
		t.Fatalf("renderBox should keep bottom border, got last line: %q", last)
	}
}

func TestTruncateLineForPaneShouldFlattenNewlinesAndTabs(t *testing.T) {
	got := truncateLineForPane("task-01\nsub\titem", 24)
	if strings.Contains(got, "\n") || strings.Contains(got, "\t") {
		t.Fatalf("truncateLineForPane should keep single line, got %q", got)
	}
	if ansi.Strip(got) != "task-01 sub item" {
		t.Fatalf("truncateLineForPane strip = %q, want %q", ansi.Strip(got), "task-01 sub item")
	}
}

func TestHelpModalShouldAvoidInlineResetArtifacts(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	modal := renderHelpModal(100)
	inlineResetPattern := regexp.MustCompile(`\x1b\[0m +[[:alpha:]]`)
	if inlineResetPattern.MatchString(modal) {
		t.Fatalf("help modal contains inline reset artifacts: %q", modal)
	}
}

func TestHelpModalTitleShouldBeHighlighted(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	modal := renderHelpModal(100)
	if !strings.Contains(modal, "HELP") {
		t.Fatalf("help modal should contain title, modal=%q", modal)
	}
	if !containsANSIColor(modal, "52;211;153") {
		t.Fatalf("help modal title should be highlighted, modal=%q", modal)
	}
}

func TestHelpModalSectionTitleShouldBeHighlighted(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	modal := renderHelpModal(100)
	if !strings.Contains(modal, "Navigation") || !strings.Contains(modal, "Task") {
		t.Fatalf("help modal should contain section titles, modal=%q", modal)
	}
	if !containsANSIColor(modal, "150;185;216") {
		t.Fatalf("help modal section title should be highlighted, modal=%q", modal)
	}
}

func TestHelpModalHintShouldUseMutedColor(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(oldProfile)
	})

	modal := renderHelpModal(100)
	if !strings.Contains(modal, "press ? / q / esc to close") {
		t.Fatalf("help modal should contain close hint, modal=%q", modal)
	}
	if !containsANSIColor(modal, "147;161;176") {
		t.Fatalf("help modal hint should use muted color, modal=%q", modal)
	}
}

func TestHelpModalShouldContainTrashActions(t *testing.T) {
	modal := ansi.Strip(renderHelpModal(100))
	if !strings.Contains(modal, "r") || !strings.Contains(modal, "restore selected in trash") {
		t.Fatalf("help modal should contain trash restore action, modal=%q", modal)
	}
	if !strings.Contains(modal, "X") || !strings.Contains(modal, "purge all in trash") {
		t.Fatalf("help modal should contain trash purge action, modal=%q", modal)
	}
}

func TestRenderTaskLineShouldAlignMetaColumnByStatusWidth(t *testing.T) {
	loc := time.Local
	due := time.Date(2026, 2, 24, 9, 30, 0, 0, loc)
	todoTask := domain.Task{Title: "todo-item", Status: domain.StatusTodo, DueAt: &due}
	doingTask := domain.Task{Title: "doing-item", Status: domain.StatusDoing, DueAt: &due}

	lineTodo := renderTaskLine("  ", renderStatusLabel(todoTask.Status), todoTask, domain.ViewInbox, loc, 120)
	lineDoing := renderTaskLine("  ", renderStatusLabel(doingTask.Status), doingTask, domain.ViewInbox, loc, 120)

	idxTodo := strings.Index(lineTodo, "due:")
	idxDoing := strings.Index(lineDoing, "due:")
	if idxTodo < 0 || idxDoing < 0 {
		t.Fatalf("both lines should contain due meta, todo=%q doing=%q", lineTodo, lineDoing)
	}
	if idxTodo != idxDoing {
		t.Fatalf("meta column should align, idxTodo=%d idxDoing=%d todo=%q doing=%q", idxTodo, idxDoing, lineTodo, lineDoing)
	}
}

func containsANSIColor(s string, colorRGB string) bool {
	return strings.Contains(s, ";38;2;"+colorRGB+"m") || strings.Contains(s, "[38;2;"+colorRGB+"m") || strings.Contains(s, ";48;2;"+colorRGB+"m") || strings.Contains(s, "[48;2;"+colorRGB+"m")
}

func containsHighlightedLabel(s, label string) bool {
	return strings.Contains(s, "[38;2;52;211;153m"+label) || strings.Contains(s, ";38;2;52;211;153m"+label)
}
