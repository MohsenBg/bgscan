package progress

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// UpdateProgressMsg updates a progress bar. Progress is normalized to [0.0, 1.0].
type UpdateProgressMsg struct {
	ID       ui.ComponentID
	Progress float64
}

func (m UpdateProgressMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}
