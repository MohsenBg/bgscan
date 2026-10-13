package confirm

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/lipgloss/v2"
)

const maxButtonGap = 20

// View renders the confirmation dialog.
func (m *Model) View() string {
	noBtn, yesBtn := m.renderButtons()
	buttons := m.layoutButtons(noBtn, yesBtn)

	message := m.messageStyle(lipgloss.Width(buttons)).Render(m.message)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		message,
		buttons,
	)
}

// renderButtons renders the No and Yes buttons with the current selection highlighted.
func (m *Model) renderButtons() (string, string) {
	no := m.buttonStyle().Render("No")
	yes := m.buttonStyle().Render("Yes")

	if m.confirm {
		yes = m.selectButtonStyle().Render("Yes")
	} else {
		no = m.selectButtonStyle().Render("No")
	}

	return no, yes
}

// layoutButtons lays out the confirmation buttons with dynamic spacing.
func (m *Model) layoutButtons(noBtn, yesBtn string) string {
	width := 80
	if w := m.env.Layout.Terminal().Width; w > 0 {
		width = w
	}

	available := width -
		lipgloss.Width(noBtn) -
		lipgloss.Width(yesBtn)

	gap := min(max(available, 1), maxButtonGap)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		noBtn,
		strings.Repeat(" ", gap),
		yesBtn,
	)
}

// Mode returns KeyModeNormal because confirmations handle their own keys.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
