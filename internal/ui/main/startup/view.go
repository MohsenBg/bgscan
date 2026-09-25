package startup

import (
	"charm.land/lipgloss/v2"
)

func (m *model) renderTitle() string {
	title := m.titleTextStyle(m.width).Render("Startup Health Check")
	bar := m.statusIndicator(category{status: m.overallStatus(), started: true}) + " " + title
	return m.titleBarStyle(m.width).Render(bar)
}

func (m *model) renderSidebar() string {
	items := make([]string, 0, len(m.categories))

	for _, cat := range m.categories {
		style := m.sidebarItemStyle()
		if cat.status == catRunning && cat.started {
			style = m.sidebarItemActiveStyle()
		}
		items = append(items, style.Render(m.statusIndicator(cat)+" "+cat.label))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *model) renderContent() string {
	contentWidth := m.viewport.Width()

	sections := make([]string, 0, len(m.categories))

	for _, cat := range m.categories {
		labelStyle := m.categoryLabelStyle()
		label := cat.label

		switch {
		case cat.status == catRunning && cat.started:
			labelStyle = m.categoryLabelActiveStyle()
			label = m.spinner.View() + " " + label
		case cat.status != catRunning:
			labelStyle = m.categoryLabelDoneStyle(cat.status)
		}

		lines := []string{labelStyle.Render(label)}

		for _, line := range cat.lines {
			lines = append(lines, m.categoryLineStyle(contentWidth).Render(line))
		}

		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	return m.contentContainerStyle(contentWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

func (m *model) syncViewport() {
	atBottom := m.viewport.AtBottom()

	m.viewport.SetContent(m.renderContent())

	if atBottom {
		m.viewport.GotoBottom()
	}
}

func (m *model) renderHelpHint() string {
	hint := "↑/↓ scroll  •  enter continue  •  ? help  •  q quit"
	return m.helpHintStyle(m.width).Render(hint)
}

func (m *model) helpView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.helpTitleStyle().Render("Help"),
		"",
		m.keyStyle().Render("↑ / ↓")+"      "+m.descStyle().Render("scroll viewport"),
		m.keyStyle().Render("enter")+"      "+m.descStyle().Render("continue"),
		m.keyStyle().Render("q")+"          "+m.descStyle().Render("exit"),
		"",
		m.descStyle().Render("press any key to close"),
	)

	return m.helpOverlayStyle().Render(content)
}

func (m *model) fatalView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.fatalTitleStyle().Render("✕ Critical Error"),
		"",
		m.fatalCategoryStyle().Render(m.fatal.category),
		m.fatalMessageStyle(fatalContentWidth(m.width)).Render(m.fatal.message),
		"",
		m.descStyle().Render("the app cannot continue — press any key to exit"),
	)

	return m.fatalOverlayStyle().Render(content)
}

func (m *model) View() string {
	if m.fatal != nil {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.fatalView(),
		)
	}

	if m.showHelp {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.helpView(),
		)
	}

	title := m.renderTitle()
	progress := m.progressBar.View()

	sidebar := m.sidebarContainerStyle(m.viewport.Height()).Render(m.renderSidebar())

	content := m.contentPaddingStyle().Render(m.viewport.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	return lipgloss.JoinVertical(lipgloss.Left, title, progress, body, m.renderHelpHint())
}
