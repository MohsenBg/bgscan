package help

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if _, ok := msg.(theme.ThemeChanged); ok {
		m.rebuild()
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize()

	case tea.KeyMsg:
		if m.mode == ShortHelp {
			if m.env.Keys.Global.Help.Matches(msg) {
				return m, m.openDialogFullHelp
			}
			return m, nil
		}

		if m.env.Keys.Global.Help.Matches(msg) {
			return m, m.closeDialogFullHelp
		}
	}

	if m.mode == LongHelp {
		before := m.vp.YOffset()

		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)

		if m.vp.YOffset() != before {
			m.dirty = true
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleResize() {
	if m.mode == LongHelp {
		m.updateViewportSize() // marks dirty
		return
	}

	if !m.shortWidthAuto {
		return
	}

	w := max(1, m.env.Layout.Body().Width)
	if w != m.shortHelpWidth {
		m.shortHelpWidth = w
		m.dirty = true
	}
}
