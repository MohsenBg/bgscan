package picker

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	width := pickerWidth(m.env)
	height := pickerHeight(m.env)

	currentDir := m.currentDirStyle(width).Render(m.FilePicker.CurrentDirectory)
	content := lipgloss.JoinVertical(lipgloss.Top, currentDir, m.FilePicker.View())

	return containerStyle(width, height).Render(content)
}
