package entry

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	env := m.session.Env()
	term := env.Layout.Body()
	banner := bannerStyle(env, term.Width, bannerHeight).Render(bannerStr)
	menu := menuStyle(term.Width, term.Height-bannerHeight).Render(m.menu.View())

	return lipgloss.JoinVertical(
		lipgloss.Center,
		banner,
		menu,
	)
}
