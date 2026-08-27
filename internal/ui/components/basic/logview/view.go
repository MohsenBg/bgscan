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

	// Title
	title := container.Render(
		TitleStyle(m.viewport.Width()).Render(m.title),
	)

	// Content with scroll bar
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

	// Help bar
	help := container.Render(
		helpStyle(m.viewport.Width()).Render(helpView()),
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
		content = BorderStyle(m.viewport.Width()).Render(content)
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
			b.WriteString(ScrollBarThumbStyle().Render("█"))
		} else {
			b.WriteString(ScrollBarStyle().Render("│"))
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// helpView renders the keyboard help bar.
func helpView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,

		helpKeyStyle().Render("↑ ↓"),
		" move  ",

		helpKeyStyle().Render("b/esc"),
		" close",
	)
}
