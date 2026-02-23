package clipboard

import (
	"strings"
	"testing"
)

func TestClipboardNormalize(t *testing.T) {
	input := "   \nTitle line\r\n\r\nhttps://example.com/a\n\nsecond line\n"

	out, truncated := Normalize(input)
	if truncated {
		t.Fatalf("unexpected truncated")
	}
	if !strings.Contains(out, "Title line") {
		t.Fatalf("normalized text should contain title, got %q", out)
	}
	if !strings.Contains(out, "https://example.com/a") {
		t.Fatalf("normalized text should keep url, got %q", out)
	}
	links := ExtractLinks(out)
	if len(links) != 1 {
		t.Fatalf("links len = %d, want 1", len(links))
	}
}
