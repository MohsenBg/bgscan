package picker

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/lipgloss/v2"
)

// pickerHeight returns the vertical space available for the picker overlay.
func pickerHeight(env *ui.Env) int {
	return env.Layout.Body().Height - 10
}

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Left, lipgloss.Top).
		Padding(0, 1).
		Margin(1, 0)
}

func (m *Model) titleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(m.env.Theme.Info).
		Padding(0, 0, 2, 0).
		BorderForeground(m.env.Theme.Border)
}

func (m *Model) currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Left).
		Bold(true).
		Foreground(m.env.Theme.Yellow).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}

func (m *Model) helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 5).
		Foreground(m.env.Theme.Muted).
		Align(lipgloss.Center).
		PaddingTop(1)
}

func (m *Model) helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.env.Theme.Secondary).
		Bold(true)
}
