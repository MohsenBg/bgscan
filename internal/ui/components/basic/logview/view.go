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

	title := container.Render(
		m.titleStyle(m.viewport.Width()).Render(m.title),
	)

	help := container.Render(
		m.helpStyle(m.viewport.Width()).Render(m.help.View()),
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
	if len(m.entries) == 0 {
		return m.loadingStyle().Render("Loading Content...!")
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

	thumb := m.scrollBarThumbStyle().Render("█")
	track := m.scrollBarStyle().Render("│")

	var b strings.Builder
	b.Grow(height * (lipgloss.Width(thumb) + 1))
	for i := range height {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			b.WriteString(thumb)
		} else {
			b.WriteString(track)
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
