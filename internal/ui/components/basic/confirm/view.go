package confirm

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/shared/env"

	"charm.land/lipgloss/v2"
)

const maxButtonGap = 20

// View renders the confirmation dialog: the message and the "No"/"Yes" buttons,
// with the focused button highlighted. Buttons are spaced by available width.
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

// renderButtons builds the styled "No" and "Yes" buttons.
//
// The button corresponding to the current selection state is rendered
// using selectButtonStyle while the other uses the normal buttonStyle.
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

// layoutButtons arranges the confirmation buttons horizontally.
//
// The spacing between buttons is dynamically calculated based on the
// terminal width while being clamped between 1 and maxButtonGap to
// prevent excessive spacing on large terminals.
func (m *Model) layoutButtons(noBtn, yesBtn string) string {
	width := 80
	if m.layout != nil && m.layout.Terminal.Width > 0 {
		width = m.layout.Terminal.Width
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

// Mode returns NormalMode because confirmations handle their own keys.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
