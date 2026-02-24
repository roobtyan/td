package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"td/internal/clipboard"
	"td/internal/domain"
)

const (
	headerLabelFG = "\x1b[38;2;52;211;153m"
	headerBaseFG  = "\x1b[38;2;217;226;236m"
	footerLabelFG = "\x1b[1;38;2;52;211;153m"
	footerBaseFG  = "\x1b[38;2;147;161;176m"
)

func bodyPaneWidths(totalWidth int) (int, int, int) {
	if totalWidth <= 0 {
		totalWidth = 80
	}

	gap := 2
	navWidth := totalWidth / 4
	if navWidth < 24 {
		navWidth = 24
	}
	if navWidth > 30 {
		navWidth = 30
	}

	listWidth := totalWidth - navWidth - gap
	if listWidth < 28 {
		listWidth = 28
		navWidth = totalWidth - listWidth - gap
	}
	if navWidth < 16 {
		navWidth = 16
		listWidth = totalWidth - navWidth - gap
	}
	if listWidth < 1 {
		listWidth = 1
	}
	return navWidth, listWidth, gap
}

func joinSections(header, body, footer string) string {
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func joinColumns(left, right []string, leftWidth, rightWidth, gap int) string {
	if gap < 1 {
		gap = 1
	}

	maxLines := len(left)
	if len(right) > maxLines {
		maxLines = len(right)
	}
	var b strings.Builder
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}
		b.WriteString(padRight(leftLine, leftWidth))
		b.WriteString(strings.Repeat(" ", gap))
		b.WriteString(padRight(rightLine, rightWidth))
		if i != maxLines-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func padRight(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if ansi.StringWidth(text) > width {
		tail := ""
		if width > 3 {
			tail = "..."
		}
		text = ansi.Truncate(text, width, tail)
	}

	displayWidth := ansi.StringWidth(text)
	if displayWidth >= width {
		return text
	}
	return text + strings.Repeat(" ", width-displayWidth)
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func splitRendered(block string) []string {
	return strings.Split(block, "\n")
}

func fitViewport(view string, width, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	lines := strings.Split(view, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = padRight(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func fitPaneHeight(lines []string, height int) []string {
	if height <= 0 {
		return nil
	}
	if len(lines) > height {
		return lines[:height]
	}
	out := make([]string, len(lines), height)
	copy(out, lines)
	for len(out) < height {
		out = append(out, "")
	}
	return out
}

func renderHeader(done, total int, view domain.View, count int, doing, todo, doneCount, overdue int, now time.Time, width int) string {
	progressBar := renderProgressBar(done, total, 18)
	percent := progressPercent(done, total)
	progressLine := fmt.Sprintf(
		"Today Progress %s %d/%d (%d%%)",
		progressBar,
		done,
		total,
		percent,
	)
	timeLine := fmt.Sprintf("Time: %s", now.Format("2006-01-02 Mon 15:04"))
	viewLine := fmt.Sprintf("View: %s  Items: %d", viewLabel(view), count)
	metricLine := fmt.Sprintf("Doing: %d  Todo: %d  Done: %d  Overdue: %d", doing, todo, doneCount, overdue)

	logoLines := []string{
		" _____ _____ ",
		"|_   _|  _  |",
		"  | | | | | |",
		"  |_| |_| |_|",
		"      TD     ",
	}
	logoRaw := strings.Join(logoLines, "\n")
	logo := logoRaw
	maxInfoWidth := width - 4 - ansi.StringWidth(logoLines[0]) - 2
	if maxInfoWidth < 10 {
		maxInfoWidth = 10
	}
	progressLine = ansi.Truncate(progressLine, maxInfoWidth, "")
	timeLine = ansi.Truncate(timeLine, maxInfoWidth, "")
	viewLine = ansi.Truncate(viewLine, maxInfoWidth, "")
	metricLine = ansi.Truncate(metricLine, maxInfoWidth, "")
	progressLine = highlightHeaderLabels(progressLine, "Today Progress")
	timeLine = highlightHeaderLabels(timeLine, "Time:")
	viewLine = highlightHeaderLabels(viewLine, "View:", "Items:")
	metricLine = highlightHeaderLabels(metricLine, "Doing:", "Todo:", "Done:", "Overdue:")
	info := progressLine + "\n" + timeLine + "\n" + viewLine + "\n" + metricLine
	content := lipgloss.JoinHorizontal(lipgloss.Top, logo, "  ", info)
	return renderBox(headerBoxStyle, content, width, 0)
}

func renderFooter(statusMsg string, focused focusArea, width int, view domain.View) string {
	status := statusMsg
	if status == "" {
		status = "ready"
	}
	focusLabel := "nav"
	if focused == focusList {
		focusLabel = "list"
	}
	content := "{status} " + status + "  {focus} " + focusLabel + "  {help} ?  {quit} q"
	if view == domain.ViewTrash {
		content += "  {restore} r  {purge} X"
	}
	maxContentWidth := width - 4
	if maxContentWidth < 0 {
		maxContentWidth = 0
	}
	content = ansi.Truncate(content, maxContentWidth, "")
	content = strings.ReplaceAll(content, "{status}", "status")
	content = strings.ReplaceAll(content, "{focus}", "focus")
	content = strings.ReplaceAll(content, "{help}", "help")
	content = strings.ReplaceAll(content, "{quit}", "quit")
	content = strings.ReplaceAll(content, "{restore}", "restore")
	content = strings.ReplaceAll(content, "{purge}", "purge")
	content = highlightFooterLabels(content, "status", "focus", "help", "quit", "restore", "purge")
	return renderBox(footerBoxStyle, content, width, 0)
}

func renderInputFooter(prompt string, width int) string {
	content := strings.TrimSpace(prompt)
	if content == "" {
		content = "input>"
	}
	maxContentWidth := width - 4
	if maxContentWidth < 0 {
		maxContentWidth = 0
	}
	content = ansi.Truncate(content, maxContentWidth, "")
	return renderBox(footerBoxStyle, content, width, 0)
}

func renderHelpModal(width int) string {
	modalWidth := width - 10
	if modalWidth > 88 {
		modalWidth = 88
	}
	if modalWidth < 48 {
		modalWidth = width - 4
	}
	if modalWidth < 34 {
		modalWidth = 34
	}

	lines := []string{
		helpTitleStyle.Render("HELP"),
		"",
		helpSectionStyle.Render("Navigation"),
		renderHelpLine("j/k, up/down", "move cursor"),
		renderHelpLine("Tab", "switch focus nav/list"),
		renderHelpLine("Enter", "select current view/project"),
		"",
		helpSectionStyle.Render("Task"),
		renderHelpLine("a", "add task"),
		renderHelpLine("e", "edit title"),
		renderHelpLine("x", "delete task / project"),
		renderHelpLine("c", "mark done"),
		renderHelpLine("z", "undo last action"),
		renderHelpLine("P", "set project"),
		renderHelpLine("t", "mark today"),
		renderHelpLine("d", "set due"),
		renderHelpLine("h", "toggle done in project"),
		renderHelpLine("r", "restore selected in trash"),
		renderHelpLine("X", "purge all in trash"),
		renderHelpLine("Space", "ai input + preview"),
		renderHelpLine("p / Ctrl+a", "ai parse clipboard"),
		"",
		helpHintStyle.Render("press ? / q / esc to close"),
	}
	return renderBox(helpModalBoxStyle, joinLines(lines), modalWidth, 0)
}

func renderAIInputModal(width int, input string, cursor int) string {
	modalWidth := width - 12
	if modalWidth > 90 {
		modalWidth = 90
	}
	if modalWidth < 48 {
		modalWidth = width - 4
	}
	if modalWidth < 36 {
		modalWidth = 36
	}

	line := renderCursorAt(input, cursor)
	lines := []string{
		helpTitleStyle.Render("AI QUICK INPUT"),
		"",
		renderHelpLine("text", truncateLineForPane(line, modalWidth-8)),
		"",
		helpHintStyle.Render("Enter preview  esc cancel"),
	}
	return renderBox(helpModalBoxStyle, joinLines(lines), modalWidth, 0)
}

func renderAIPreviewModal(width int, parsed clipboard.ParsedTask, source string) string {
	modalWidth := width - 12
	if modalWidth > 92 {
		modalWidth = 92
	}
	if modalWidth < 52 {
		modalWidth = width - 4
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	sourceLabel := "Fallback"
	if strings.TrimSpace(strings.ToLower(source)) == "ai" {
		sourceLabel = "AI"
	}
	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		title = "-"
	}
	project := strings.TrimSpace(parsed.Project)
	if project == "" {
		project = "-"
	}
	due := strings.TrimSpace(parsed.Due)
	if due == "" {
		due = "-"
	}
	priority := strings.TrimSpace(parsed.Priority)
	if priority == "" {
		priority = "P2"
	}

	lines := []string{
		helpTitleStyle.Render("AI PREVIEW"),
		"",
		renderHelpLine("source", sourceLabel),
		renderHelpLine("todo", truncateLineForPane(title, modalWidth-10)),
		renderHelpLine("project", truncateLineForPane(project, modalWidth-10)),
		renderHelpLine("due", due),
		renderHelpLine("priority", priority),
		"",
		helpHintStyle.Render("Enter confirm  e edit  esc cancel"),
	}
	return renderBox(helpModalBoxStyle, joinLines(lines), modalWidth, 0)
}

func renderHelpLine(keys, desc string) string {
	return padRight(keys, 14) + " " + desc
}

func renderCursorAt(text string, cursor int) string {
	runes := []rune(text)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	out := make([]rune, 0, len(runes)+1)
	out = append(out, runes[:cursor]...)
	out = append(out, '|')
	out = append(out, runes[cursor:]...)
	return string(out)
}

func renderDimmedPage(page string, width, height int) string {
	base := fitViewport(ansi.Strip(page), width, height)
	lines := strings.Split(base, "\n")
	for i := range lines {
		lines[i] = helpBackdropLineStyle.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}

func overlayCentered(base, block string, width, height int) string {
	blockLines := strings.Split(block, "\n")
	blockHeight := len(blockLines)
	blockWidth := 0
	for _, line := range blockLines {
		w := ansi.StringWidth(line)
		if w > blockWidth {
			blockWidth = w
		}
	}
	x := (width - blockWidth) / 2
	if x < 0 {
		x = 0
	}
	y := (height - blockHeight) / 2
	if y < 0 {
		y = 0
	}

	baseLines := strings.Split(fitViewport(base, width, height), "\n")
	for i, blockLine := range blockLines {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		if x >= width {
			continue
		}
		line := blockLine
		available := width - x
		if available <= 0 {
			continue
		}
		if ansi.StringWidth(line) > available {
			line = ansi.Truncate(line, available, "")
		}
		lineWidth := ansi.StringWidth(line)
		if lineWidth <= 0 {
			continue
		}
		left := ansi.Cut(baseLines[row], 0, x)
		right := ansi.Cut(baseLines[row], x+lineWidth, width)
		baseLines[row] = padRight(left, x) + line + right
		baseLines[row] = padRight(baseLines[row], width)
	}
	return strings.Join(baseLines, "\n")
}

func renderBox(style lipgloss.Style, content string, width, height int) string {
	s := style.Copy()
	styleWidth := width - s.GetHorizontalBorderSize() - s.GetHorizontalMargins()
	if styleWidth < 0 {
		styleWidth = 0
	}
	contentWidth := styleWidth - s.GetHorizontalPadding()
	if contentWidth < 0 {
		contentWidth = 0
	}

	contentHeight := -1
	s = s.Width(styleWidth)
	if height > 0 {
		styleHeight := height - s.GetVerticalBorderSize() - s.GetVerticalMargins()
		if styleHeight < 0 {
			styleHeight = 0
		}
		contentHeight = styleHeight - s.GetVerticalPadding()
		if contentHeight < 0 {
			contentHeight = 0
		}
		s = s.Height(styleHeight)
	}
	content = fitContentForBox(content, contentWidth, contentHeight)
	return s.Render(content)
}

func progressPercent(done, total int) int {
	if total <= 0 {
		return 100
	}
	if done <= 0 {
		return 0
	}
	return int(float64(done) / float64(total) * 100)
}

func renderProgressBar(done, total, width int) string {
	if width <= 0 {
		width = 12
	}
	fill := 0
	if total > 0 && done > 0 {
		fill = int(float64(done) / float64(total) * float64(width))
	}
	if fill > width {
		fill = width
	}
	if fill < 0 {
		fill = 0
	}
	return "[" + strings.Repeat("#", fill) + strings.Repeat("-", width-fill) + "]"
}

func viewLabel(v domain.View) string {
	switch v {
	case domain.ViewToday:
		return "Today"
	case domain.ViewInbox:
		return "Inbox"
	case domain.ViewLog:
		return "Log"
	case domain.ViewProject:
		return "Project"
	case domain.ViewTrash:
		return "Trash"
	default:
		return string(v)
	}
}

func highlightHeaderLabels(line string, labels ...string) string {
	out := line
	for _, label := range labels {
		out = strings.ReplaceAll(out, label, headerLabelFG+label+headerBaseFG)
	}
	return out
}

func highlightFooterLabels(line string, labels ...string) string {
	out := line
	for _, label := range labels {
		out = strings.ReplaceAll(out, label, footerLabelFG+label+footerBaseFG)
	}
	return out
}

func paneContentWidth(width int) int {
	w := width - 6
	if w < 1 {
		return 1
	}
	return w
}

func paneContentHeight(height int) int {
	h := height - 4
	if h < 1 {
		return 1
	}
	return h
}

func viewportWindow(total, cursor, visible int) (int, int) {
	if total <= 0 || visible <= 0 {
		return 0, 0
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	if total <= visible {
		return 0, total
	}
	start := 0
	if cursor >= visible {
		start = cursor - visible + 1
	}
	maxStart := total - visible
	if start > maxStart {
		start = maxStart
	}
	end := start + visible
	if end > total {
		end = total
	}
	return start, end
}

func truncateLineForPane(line string, width int) string {
	line = flattenPaneLine(line)
	if width <= 0 || ansi.StringWidth(line) <= width {
		return line
	}
	tail := ""
	if width > 3 {
		tail = "..."
	}
	return ansi.Truncate(line, width, tail)
}

func fitContentForBox(content string, width, maxLines int) string {
	if width <= 0 || maxLines == 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if width > 0 {
		for i := range lines {
			lines[i] = ansi.Truncate(lines[i], width, "")
		}
	}
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

func flattenPaneLine(line string) string {
	line = strings.ReplaceAll(line, "\r\n", " ")
	line = strings.ReplaceAll(line, "\n", " ")
	line = strings.ReplaceAll(line, "\r", " ")
	line = strings.ReplaceAll(line, "\t", " ")
	return line
}
