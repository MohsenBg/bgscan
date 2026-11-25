package theme

import tea "charm.land/bubbletea/v2"

// ThemeChanged is the message broadcast after the active theme
// changes, so components that cache styles can rebuild them.
type ThemeChanged struct{}

// ThemeChangedCmd returns a command emitting ThemeChanged.
func ThemeChangedCmd() tea.Cmd {
	return func() tea.Msg {
		return ThemeChanged{}
	}
}
