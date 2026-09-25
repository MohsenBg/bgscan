package picker

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"

	"charm.land/lipgloss/v2"
)

// pickerHeight returns the vertical space available for the picker overlay.
func pickerHeight(layout *layout.Layout) int {
	return layout.Body.Height - 10
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
		Foreground(m.theme.Info).
		Padding(0, 0, 2, 0).
		BorderForeground(m.theme.Border)
}

func (m *Model) currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Left).
		Bold(true).
		Foreground(m.theme.Yellow).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}

func (m *Model) helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 5).
		Foreground(m.theme.Muted).
		Align(lipgloss.Center).
		PaddingTop(1)
}

func (m *Model) helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Bold(true)
}
