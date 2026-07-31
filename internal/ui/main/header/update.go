package header

import (
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update is a no-op; the header does not process messages.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	return m, nil
}
