package ipviewer

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestKeepRowBackground(t *testing.T) {
	v := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D787")).Render("25.31ms")

	kept := keepRowBackground(v)
	if strings.HasSuffix(kept, ansi.ResetStyle) || strings.HasSuffix(kept, "\x1b[0m") {
		t.Fatalf("trailing full reset must be gone: %q", kept)
	}
	if !strings.HasSuffix(kept, "\x1b[39m") {
		t.Fatalf("want SGR 39 (default fg) tail: %q", kept)
	}

	// Inside a selected-row wrapper the background must survive the colored
	// cell: exactly one reset remains, at the very end of the row.
	outer := lipgloss.NewStyle().
		Background(lipgloss.Color("#303030")).
		Render("pre" + kept + "post")

	resets := strings.Count(outer, ansi.ResetStyle) + strings.Count(outer, "\x1b[0m")
	if resets != 1 {
		t.Fatalf("row contains %d resets, want exactly 1 at the end: %q", resets, outer)
	}
	if !strings.Contains(outer, ansi.ResetStyle+"\x1b[39m") {
		// 39 restores default fg mid-cell; the wrapper reset lands last.
		if !strings.HasSuffix(outer, ansi.ResetStyle) {
			t.Fatalf("final reset must terminate the row: %q", outer)
		}
	}
}
