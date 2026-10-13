package startup

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

func (m *model) titleBarStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		BorderForeground(m.session.Env().Theme.BorderActive)
}

func (m *model) titleTextStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(m.session.Env().Theme.Text)
}

func (m *model) sidebarContainerStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, true, false, false).
		BorderForeground(m.session.Env().Theme.BorderActive)
}

func (m *model) sidebarItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Width(sidebarWidth - 2)
}

func (m *model) sidebarItemActiveStyle() lipgloss.Style {
	return m.sidebarItemStyle().
		Foreground(m.session.Env().Theme.Primary).
		Bold(true)
}

func (m *model) contentContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func (m *model) contentPaddingStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

func (m *model) helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Height(helpHeight).
		Align(lipgloss.Center, lipgloss.Bottom)
}

func (m *model) categoryLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.session.Env().Theme.Text)
}

func (m *model) categoryLabelActiveStyle() lipgloss.Style {
	return m.categoryLabelStyle().
		Foreground(m.session.Env().Theme.Primary)
}

func (m *model) categoryLabelDoneStyle(status categoryStatus) lipgloss.Style {
	return m.categoryLabelStyle().
		Foreground(m.statusColor(status))
}

func (m *model) categoryLineStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		Width(width).
		Foreground(m.session.Env().Theme.Muted)
}

func (m *model) statusColor(status categoryStatus) color.Color {
	switch status {
	case catOK:
		return m.session.Env().Theme.Success
	case catWarn:
		return m.session.Env().Theme.Yellow
	case catError:
		return m.session.Env().Theme.Error
	case catWait:
		return m.session.Env().Theme.Primary
	default:
		return m.session.Env().Theme.Info
	}
}

func (m *model) statusPrefixStyle(status categoryStatus) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.statusColor(status))
}

func (m *model) spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.session.Env().Theme.Primary)
}

func (m *model) pendingDotStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.session.Env().Theme.Muted)
}

func (m *model) helpHintStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 0).
		Align(lipgloss.Center).
		Foreground(m.session.Env().Theme.Muted).
		Faint(true)
}

func (m *model) helpOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.session.Env().Theme.Text).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.session.Env().Theme.BorderActive).
		Padding(1, 2)
}

func (m *model) descStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.session.Env().Theme.Muted)
}

func (m *model) fatalOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.session.Env().Theme.Text).
		Border(lipgloss.ThickBorder()).
		BorderForeground(m.session.Env().Theme.Error).
		Padding(1, 3)
}

func (m *model) fatalTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.session.Env().Theme.Error)
}

func (m *model) fatalCategoryStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(m.session.Env().Theme.Text)
}

func (m *model) fatalMessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Foreground(m.session.Env().Theme.Muted)
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
