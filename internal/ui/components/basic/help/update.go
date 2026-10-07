package help

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize()

	case tea.KeyMsg:
		if m.mode == ShortHelp {
			if m.deps.Keys.Global.Help.Matches(msg) {
				return m, m.openDialogFullHelp
			}

			return m, nil
		}

		if m.mode == LongHelp {
			if m.deps.Keys.Global.Help.Matches(msg) {
				return m, m.closeDialogFullHelp
			}
		}
	}

	if m.mode == LongHelp {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleResize() {
	if m.mode == LongHelp {
		m.updateViewportSize()
		return
	}

	if m.shortWidthAuto {
		m.shortHelpWidth = max(1, m.deps.Layout.BodyContentWidth())
	}
}
