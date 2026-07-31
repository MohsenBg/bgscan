package progress

import (
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

// Update resizes the bar on terminal resize, advances the animation, and
// applies UpdateProgressMsg to set the current percentage.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	// Terminal resized → recompute progress width
	case tea.WindowSizeMsg:
		m.progress.SetWidth(m.Width())
		return m, nil

	// External progress update
	case UpdateProgressMsg:
		if msg.ID != m.ID() {
			return m, nil
		}
		// When progress reaches completion, switch to a cleaner format
		if msg.Progress >= 1 {
			m.progress.PercentFormat = " %3.0f%%"
			return m, m.progress.SetPercent(1)
		}

		m.progress.PercentFormat = " %0.2f%%"
		return m, m.progress.SetPercent(msg.Progress)

	// progress.FrameMsg is emitted internally by the progress model
	// to animate the bar smoothly.
	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model
		return m, cmd
	}

	return m, nil
}
