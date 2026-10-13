package theme

import tea "charm.land/bubbletea/v2"

type ThemeChanged struct{}

// ThemeChangedCmd returns a command emitting ThemeChanged.
func ThemeChangedCmd() tea.Cmd {
	return func() tea.Msg {
		return ThemeChanged{}
	}
}
