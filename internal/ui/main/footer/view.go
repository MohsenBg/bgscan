package footer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// View renders the footer component, displaying application
// information, current status, and runtime metrics.
func (m *Model) View() string {
	width := m.layout.Footer.Width
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
	centerSection := centerSectionStyle(centerWidth - 3).Render(
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

	rightSection := rightSectionStyle(rightWidth + 3).Render(runtimeInfo)

	// Assemble footer content horizontally
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
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
