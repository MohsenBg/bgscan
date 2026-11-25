package dnstt

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update forwards the message to the shared form base and returns the
// model itself.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	return m.Base.Update(m, msg)
}
