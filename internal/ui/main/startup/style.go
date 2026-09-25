package startup

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

func (m *model) titleBarStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		BorderForeground(m.state.Theme().BorderActive)
}

func (m *model) titleTextStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(m.state.Theme().Text)
}

func (m *model) sidebarContainerStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, true, false, false).
		BorderForeground(m.state.Theme().BorderActive)
}

func (m *model) sidebarItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Width(sidebarWidth - 2)
}

func (m *model) sidebarItemActiveStyle() lipgloss.Style {
	return m.sidebarItemStyle().
		Foreground(m.state.Theme().Primary).
		Bold(true)
}

func (m *model) contentContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func (m *model) contentPaddingStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

func (m *model) categoryLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Text)
}

func (m *model) categoryLabelActiveStyle() lipgloss.Style {
	return m.categoryLabelStyle().Foreground(m.state.Theme().Primary)
}

func (m *model) categoryLabelDoneStyle(status categoryStatus) lipgloss.Style {
	return m.categoryLabelStyle().Foreground(m.statusColor(status))
}

func (m *model) categoryLineStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		Width(width).
		Foreground(m.state.Theme().Muted)
}

func (m *model) statusColor(status categoryStatus) color.Color {
	switch status {
	case catOK:
		return m.state.Theme().Success
	case catWarn:
		return m.state.Theme().Yellow
	case catError:
		return m.state.Theme().Error
	case catWait:
		return m.state.Theme().Primary
	default:
		return m.state.Theme().Info
	}
}

func (m *model) statusPrefixStyle(status categoryStatus) lipgloss.Style {
	switch status {
	case catOK:
		return lipgloss.NewStyle().Foreground(m.state.Theme().Success)
	case catWarn:
		return lipgloss.NewStyle().Foreground(m.state.Theme().Yellow)
	case catError:
		return lipgloss.NewStyle().Foreground(m.state.Theme().Error)
	case catWait:
		return lipgloss.NewStyle().Foreground(m.state.Theme().Primary)
	default:
		return lipgloss.NewStyle().Foreground(m.state.Theme().Info)
	}
}

func (m *model) spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.state.Theme().Primary)
}

func (m *model) pendingDotStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.state.Theme().Muted)
}

func (m *model) helpHintStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 0).
		Align(lipgloss.Center).
		Foreground(m.state.Theme().Muted).
		Faint(true)
}

func (m *model) helpOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.state.Theme().Text).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.state.Theme().BorderActive).
		Padding(1, 2)
}

func (m *model) helpTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Primary)
}

func (m *model) keyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.state.Theme().Info).Bold(true)
}

func (m *model) descStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.state.Theme().Muted)
}

func (m *model) fatalOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.state.Theme().Text).
		Border(lipgloss.ThickBorder()).
		BorderForeground(m.state.Theme().Error).
		Padding(1, 3)
}

func (m *model) fatalTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Error)
}

func (m *model) fatalCategoryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.state.Theme().Text)
}

func (m *model) fatalMessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Foreground(m.state.Theme().Muted)
}

func fatalContentWidth(termWidth int) int {
	w := termWidth - 20
	if w > 70 {
		w = 70
	}
	if w < 20 {
		w = 20
	}
	return w
}
