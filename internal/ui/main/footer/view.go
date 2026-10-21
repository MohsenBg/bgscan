package footer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

func (m *Model) View() string {
	term := m.env.Layout.Footer()
	th := m.env.Theme

	leftWidth := term.Width / 3
	centerWidth := term.Width / 3
	rightWidth := term.Width - leftWidth - centerWidth

	leftSection := leftSectionStyle(leftWidth).Render(
		fmt.Sprintf("%s %s",
			appNameStyle(&th).Render("⚡ BGScan"),
			versionStyle(&th).Render(m.appVersion),
		),
	)

	centerSection := centerSectionStyle(centerWidth - 2).Render(
		statusTextStyle(&th).Render(m.status),
	)

	runtimeInfo := fmt.Sprintf("%s %s %s %s",
		statsStyle(&th).Render("GR:"),
		statsValueStyle(&th).Render(fmt.Sprintf("%d", m.goroutines)),
		statsStyle(&th).Render("Mem:"),
		statsValueStyle(&th).Render(humanize.Bytes(m.memoryBytes)),
	)

	rightSection := rightSectionStyle(rightWidth + 2).Render(runtimeInfo)

	footerContent := lipgloss.JoinHorizontal(lipgloss.Left, leftSection, centerSection, rightSection)
	separator := separatorStyle(&th, term.Width).Render(strings.Repeat("─", term.Width))

	return containerStyle(&th, term.Width, term.Height).Render(
		lipgloss.JoinVertical(lipgloss.Left, separator, footerContent),
	)
}
