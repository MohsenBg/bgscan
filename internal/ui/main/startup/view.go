package startup

import (
	"charm.land/lipgloss/v2"
)

func (m *model) renderTitle() string {
	indicator := m.statusIndicator(category{
		status:  m.overallStatus(),
		started: true,
	})

	title := m.titleTextStyle(m.width).Render("Startup Health Check")

	return m.titleBarStyle(m.width).Render(indicator + " " + title)
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
	contentWidth := max(m.viewport.Width(), 1)
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

func (m *model) fatalView() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.fatalTitleStyle().Render("✕ Critical Error"),
		"",
		m.fatalCategoryStyle().Render(m.fatal.category),
		m.fatalMessageStyle(fatalContentWidth(m.width)).Render(m.fatal.message),
		"",
		m.descStyle().Render("the app cannot continue. press any key to exit"),
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

	title := m.renderTitle()
	progressBar := m.progressBar.View()

	sidebarHeight := max(m.viewport.Height(), 1)
	sidebar := m.sidebarContainerStyle(sidebarHeight).Render(m.renderSidebar())
	content := m.contentPaddingStyle().Render(m.viewport.View())
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	helpView := ""
	if m.help != nil {
		helpView = m.help.View()
	}

	help := m.helpStyle().Render(helpView)

	return lipgloss.JoinVertical(lipgloss.Left, title, progressBar, body, help)
}
