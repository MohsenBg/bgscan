package thefeed

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update handles msg through the shared form base and returns the
// updated component and any command the base produces.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	return m.Base.Update(m, msg)
}
