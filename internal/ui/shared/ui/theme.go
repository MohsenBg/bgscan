package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Themed is implemented by components that cache theme-derived styles
// (lipgloss styles, bubbles styles, delegates). Call SetTheme when the
// active theme changes.
type Themed interface {
	SetTheme(*theme.Theme)
}

// HandleTheme applies a ThemeChangedMsg to all targets.
// It returns true when msg was a ThemeChangedMsg (caller should return
// early), false otherwise. Nil targets are skipped.
func HandleTheme(msg tea.Msg, targets ...Themed) bool {
	th, ok := msg.(ThemeChangedMsg)
	if !ok {
		return false
	}
	for _, t := range targets {
		if t != nil {
			t.SetTheme(th.Theme)
		}
	}
	return true
}

// SyncState applies a ThemeChangedMsg to shared AppState.
// Used by root models (app, workspace) that own no cached styles
// themselves but must keep Settings.Theme in sync before forwarding
// the message to children. Returns true when msg was handled.
func SyncState(msg tea.Msg, s *AppState) bool {
	th, ok := msg.(ThemeChangedMsg)
	if !ok {
		return false
	}
	if s != nil {
		s.Settings.Theme = th.Theme
		s.Settings.ThemeName = th.Name
	}
	return true
}

// ThemeChangedCmd returns a command emitting ThemeChangedMsg.
// Prefer this over Program.Send inside Update/submit callbacks:
// returning a Cmd lets BubbleTea schedule the broadcast on the event
// loop instead of pushing it re-entrantly from inside Update.
func ThemeChangedCmd(th *theme.Theme, name string) tea.Cmd {
	return func() tea.Msg {
		return ThemeChangedMsg{Theme: th, Name: name}
	}
}
