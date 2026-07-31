package footer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

// View renders the footer: app name/version on the left, status in the center,
// and runtime stats on the right.
func (m *Model) View() string {
	padding := 2
	width := m.layout.Footer.Width - padding
	height := m.layout.Footer.Height

	// Divide footer into three sections
	leftWidth := width / 3
	centerWidth := width / 3
	rightWidth := width - leftWidth - centerWidth

	// Left section: application info
	leftSection := leftSectionStyle(leftWidth).Render(
		fmt.Sprintf(
			"%s %s %s",
			iconStyle().Render("⚡"),
			appNameStyle().Render("BGScan"),
			versionStyle().Render("v"+m.appVersion),
		),
	)

	// Center section: application status
	centerSection := centerSectionStyle(centerWidth - 2).Render(
		statusTextStyle().Render(m.status),
	)

	// Right section: runtime metrics
	runtimeInfo := fmt.Sprintf(
		"%s GR:%d | %s Mem:%s",
		iconStyle().Render("⚙"),
		m.goroutines,
		iconStyle().Render("🧠"),
		humanize.Bytes(m.memoryBytes),
	)

	rightSection := rightSectionStyle(rightWidth + 2).Render(runtimeInfo)

	// Assemble footer content horizontally
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftSection,
		centerSection,
		rightSection,
	)

	// Top separator line
	separator := separatorStyle(width).Render(strings.Repeat("─", width))

	// Final footer layout
	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		separator,
		footerContent,
	)

	return containerStyle(width, height).Render(footer)
}
