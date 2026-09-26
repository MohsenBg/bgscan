package footer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

func (m *Model) View() string {
	padding := 2
	width := m.layout.Footer.Width - padding
	height := m.layout.Footer.Height

	leftWidth := width / 3
	centerWidth := width / 3
	rightWidth := width - leftWidth - centerWidth

	leftSection := leftSectionStyle(leftWidth).Render(
		fmt.Sprintf("%s %s",
			m.appNameStyle().Render("⚡ BGScan"),
			m.versionStyle().Render("v"+m.appVersion),
		),
	)

	centerSection := centerSectionStyle(centerWidth - 2).Render(
		m.statusTextStyle().Render(m.status),
	)

	runtimeInfo := fmt.Sprintf("%s %s %s %s",
		m.statsStyle().Render("GR:"),
		m.statsValueStyle().Render(fmt.Sprintf("%d", m.goroutines)),
		m.statsStyle().Render("Mem:"),
		m.statsValueStyle().Render(humanize.Bytes(m.memoryBytes)),
	)

	rightSection := rightSectionStyle(rightWidth + 2).Render(runtimeInfo)

	footerContent := lipgloss.JoinHorizontal(lipgloss.Left, leftSection, centerSection, rightSection)
	separator := m.separatorStyle(width).Render(strings.Repeat("─", width))

	return m.containerStyle(width, height).Render(
		lipgloss.JoinVertical(lipgloss.Left, separator, footerContent),
	)
}
