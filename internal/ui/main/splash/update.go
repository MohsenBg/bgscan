package splash

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update advances the frame counter on each tick and ends the screen
// when the total frame count is reached or any key is pressed.
func (m model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		// Frames advance only when the terminal has room to draw the
		// logo, so a too small terminal pauses the screen instead of
		// skipping it.
		if m.session.Env().Layout.HasSpace() {
			m.frame++
		}
		if m.frame >= totalFrames {
			return m, func() tea.Msg { return SplashDoneMsg{} }
		}
		return m, tickCmd()

	case tea.KeyMsg:
		return m, func() tea.Msg { return SplashDoneMsg{} }
	}
	return m, nil
}
