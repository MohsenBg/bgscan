package menu

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Update handles resize and key input. Unhandled messages are forwarded to
// the underlying list to preserve its built-in navigation behavior.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateMenuLayout()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.GetSelected(); ok {
				if item.action != nil {
					return m, item.action()
				}
				if m.onSelect != nil {
					return m, m.onSelect(item)
				}
			}
		case "q", env.KeyEsc:
			return m, cmd
		}

		for i, l := range m.items {
			if l.shortcut == msg.String() {
				m.List.Select(i)
				if item, ok := m.GetSelected(); ok {
					if item.action != nil {
						return m, item.action()
					}
					if m.onSelect != nil {
						return m, m.onSelect(item)
					}
				}
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// updateMenuLayout clamps the menu to a maximum size so it doesn't expand
// unboundedly on large terminals while still shrinking on small ones.
func (m *Model) updateMenuLayout() {
	width := m.width
	height := m.height

	if m.widthAuto {
		width = min(m.Layout.BodyContentWidth(), 50)
	}

	if m.heightAuto {
		height = min(m.Layout.BodyContentHeight(), 20)
	}

	m.List.Styles.TitleBar = m.List.Styles.TitleBar.
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)

	m.List.SetWidth(width)
	m.List.SetHeight(height)
}
