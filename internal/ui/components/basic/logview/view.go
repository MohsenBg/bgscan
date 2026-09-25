package logview

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View renders the log viewer component.
//
// Layout:
//
//	title
//	log viewport + scroll bar
//	help bar
//
// When no log messages are available, a loading placeholder is displayed.
func (m *Model) View() string {
	container := ContainerStyle(m.containerWidth)

	title := container.Render(
		m.titleStyle(m.viewport.Width()).Render(m.title),
	)

	content := m.renderContentView()
	scrollBar := m.renderScrollBar()

	var contentArea string
	if scrollBar != "" {
		contentArea = lipgloss.JoinHorizontal(
			lipgloss.Top,
			container.Render(content),
			scrollBar,
		)
	} else {
		contentArea = container.Render(content)
	}

	help := container.Render(
		m.helpStyle(m.viewport.Width()).Render(m.helpView()),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		contentArea,
		help,
	)
}

// renderContentView renders the viewport or loading state.
func (m *Model) renderContentView() string {
	if len(m.messages) == 0 {
		return "Loading Content...!"
	}

	content := m.viewport.View()

	if m.showBorder {
		content = m.borderStyle(m.viewport.Width()).Render(content)
	}

	return content
}

// renderScrollBar renders a vertical scroll bar indicator.
func (m *Model) renderScrollBar() string {
	totalLines := m.viewport.TotalLineCount()
	visibleLines := m.viewport.VisibleLineCount()
	if totalLines <= visibleLines {
		return ""
	}
	scrollPercent := m.viewport.ScrollPercent()
	height := m.viewport.Height()

	thumbHeight := max(1, (visibleLines*height)/totalLines)
	trackSpace := height - thumbHeight
	thumbPos := int(scrollPercent * float64(trackSpace))

	var b strings.Builder
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			b.WriteString(m.scrollBarThumbStyle().Render("█"))
		} else {
			b.WriteString(m.scrollBarStyle().Render("│"))
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// helpView renders the keyboard help bar.
func (m *Model) helpView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,

		m.helpKeyStyle().Render("↑ ↓"),
		" move  ",

		m.helpKeyStyle().Render("b/esc"),
		" close",
	)
}
