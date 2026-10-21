package splash

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
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
