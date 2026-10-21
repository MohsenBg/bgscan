// style.go
package picker

import (
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

func pickerWidth(env *ui.Env) int {
	return max(40, min(70, env.Layout.Content().Width-10))
}

func pickerHeight(env *ui.Env) int {
	return max(10, min(env.Layout.Body().Height-12, 40))
}

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Left, lipgloss.Top).
		Padding(0, 1).
		Margin(1, 0)
}

func (m *Model) currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(max(10, width-5)).
		Align(lipgloss.Left).
		Bold(true).
		Foreground(m.env.Theme.Yellow).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}
