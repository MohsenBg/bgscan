package startup

import (
	"charm.land/lipgloss/v2"
)

func (m *model) renderTitle() string {
	title := titleTextStyle(m.width).Render("Startup Health Check")
	bar := m.statusIndicator(category{status: m.overallStatus(), started: true}) + " " + title
	return titleBarStyle(m.width).Render(bar)
}

func (m *model) renderSidebar() string {
	items := make([]string, 0, len(m.categories))

	for _, cat := range m.categories {
		style := sidebarItemStyle()
		if cat.status == catRunning && cat.started {
			style = sidebarItemActiveStyle()
		}
		items = append(items, style.Render(m.statusIndicator(cat)+" "+cat.label))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *model) renderContent() string {
	contentWidth := m.viewport.Width()

	sections := make([]string, 0, len(m.categories))

	for _, cat := range m.categories {
		labelStyle := categoryLabelStyle()
		label := cat.label

		switch {
		case cat.status == catRunning && cat.started:
			labelStyle = categoryLabelActiveStyle()
			label = m.spinner.View() + " " + label
		case cat.status != catRunning:
			labelStyle = categoryLabelDoneStyle(cat.status)
		}

		lines := []string{labelStyle.Render(label)}

		for _, line := range cat.lines {
			lines = append(lines, categoryLineStyle(contentWidth).Render(line))
		}

		sections = append(sections, lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	return contentContainerStyle(contentWidth).Render(
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
	return helpHintStyle(m.width).Render(hint)
}

func (m *model) helpView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		helpTitleStyle().Render("Help"),
		"",
		keyStyle().Render("↑ / ↓")+"      "+descStyle().Render("scroll viewport"),
		keyStyle().Render("enter")+"      "+descStyle().Render("continue"),
		keyStyle().Render("q")+"          "+descStyle().Render("exit"),
		"",
		descStyle().Render("press any key to close"),
	)

	return helpOverlayStyle().Render(content)
}

func (m *model) fatalView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		fatalTitleStyle().Render("✕ Critical Error"),
		"",
		fatalCategoryStyle().Render(m.fatal.category),
		fatalMessageStyle(fatalContentWidth(m.width)).Render(m.fatal.message),
		"",
		descStyle().Render("the app cannot continue — press any key to exit"),
	)

	return fatalOverlayStyle().Render(content)
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

	sidebar := sidebarContainerStyle(m.viewport.Height()).Render(m.renderSidebar())

	content := contentPaddingStyle().Render(m.viewport.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	return lipgloss.JoinVertical(lipgloss.Left, title, progress, body, m.renderHelpHint())
}
