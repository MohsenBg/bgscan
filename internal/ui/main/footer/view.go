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
		fmt.Sprintf("%s %s %s",
			iconStyle().Render("⚡"),
			appNameStyle().Render("BGScan"),
			versionStyle().Render("v"+m.appVersion),
		),
	)

	centerSection := centerSectionStyle(centerWidth - 2).Render(
		statusTextStyle().Render(m.status),
	)

	runtimeInfo := fmt.Sprintf("%s GR:%d | %s Mem:%s",
		iconStyle().Render("⚙"), m.goroutines,
		iconStyle().Render("🧠"), humanize.Bytes(m.memoryBytes),
	)

	rightSection := rightSectionStyle(rightWidth + 2).Render(runtimeInfo)

	footerContent := lipgloss.JoinHorizontal(lipgloss.Left, leftSection, centerSection, rightSection)
	separator := separatorStyle(width).Render(strings.Repeat("─", width))

	return containerStyle(width, height).Render(
		lipgloss.JoinVertical(lipgloss.Left, separator, footerContent),
	)
}
