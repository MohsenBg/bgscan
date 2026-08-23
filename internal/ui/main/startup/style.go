package startup

import (
	"image/color"

	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

func titleBarStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		BorderForeground(theme.Current().BorderActive)
}

func titleTextStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(theme.Current().Text)
}

func sidebarContainerStyle(height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder(), false, true, false, false).
		BorderForeground(theme.Current().BorderActive)
}

func sidebarItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Width(sidebarWidth - 2)
}

func sidebarItemActiveStyle() lipgloss.Style {
	return sidebarItemStyle().
		Foreground(theme.Current().Primary).
		Bold(true)
}

func contentContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func contentPaddingStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

func categoryLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Text)
}

func categoryLabelActiveStyle() lipgloss.Style {
	return categoryLabelStyle().Foreground(theme.Current().Primary)
}

func categoryLabelDoneStyle(status categoryStatus) lipgloss.Style {
	return categoryLabelStyle().Foreground(statusColor(status))
}

func categoryLineStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		Width(width).
		Foreground(theme.Current().Muted)
}

func statusColor(status categoryStatus) color.Color {
	switch status {
	case catOK:
		return theme.Current().Success
	case catWarn:
		return theme.Current().Yellow
	case catError:
		return theme.Current().Error
	case catWait:
		return theme.Current().Primary
	default:
		return theme.Current().Info
	}
}

func statusPrefixStyle(status categoryStatus) lipgloss.Style {
	switch status {
	case catOK:
		return lipgloss.NewStyle().Foreground(theme.Current().Success)
	case catWarn:
		return lipgloss.NewStyle().Foreground(theme.Current().Yellow)
	case catError:
		return lipgloss.NewStyle().Foreground(theme.Current().Error)
	case catWait:
		return lipgloss.NewStyle().Foreground(theme.Current().Primary)
	default:
		return lipgloss.NewStyle().Foreground(theme.Current().Info)
	}
}

func spinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Primary)
}

func pendingDotStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Muted)
}

func helpHintStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Padding(1, 0).
		Align(lipgloss.Center).
		Foreground(theme.Current().Muted).
		Faint(true)
}

func helpOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive).
		Padding(1, 2)
}

func helpTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Primary)
}

func keyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Info).Bold(true)
}

func descStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Current().Muted)
}

func fatalOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Border(lipgloss.ThickBorder()).
		BorderForeground(theme.Current().Error).
		Padding(1, 3)
}

func fatalTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Error)
}

func fatalCategoryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Current().Text)
}

func fatalMessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Foreground(theme.Current().Muted)
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
