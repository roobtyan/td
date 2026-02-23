package tui

import (
	"strings"
	"testing"
)

func TestLayoutRatio(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "Inbox") {
		t.Fatalf("view should contain Inbox, got: %q", view)
	}
}
