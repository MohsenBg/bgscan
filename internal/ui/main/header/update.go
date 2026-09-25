package header

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, nil
	}
	return m, nil
}
