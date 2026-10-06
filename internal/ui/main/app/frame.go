package app

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
)

const minFrameWidth = 16

// frameView renders dialog content with the title and hints in the frame.
func (m *model) frameView(content string, opts *frameOptions, hint string) string {
	innerWidth := max(minFrameWidth, m.state.Layout.Body.Width)

	body := lipgloss.NewStyle().
		MaxWidth(innerWidth).
		Padding(0, 1).
		Render(content)

	width := max(lipgloss.Width(body), minFrameWidth)

	rule, titleStyle := m.frameStyles(opts)

	title := ""
	if opts.Title != "" {
		title = titleStyle.Render(" " + opts.Title + " ")
	}

	rows := strings.Split(body, "\n")
	lines := make([]string, 0, len(rows)+2)

	lines = append(lines, frameEdge(rule, "╭", "┐", title, width+2, false))

	for _, row := range rows {
		lines = append(lines, rule.Render("│")+padTo(row, width)+rule.Render("│"))
	}

	hintText := ""
	if hint != "" {
		hintText = " " + hint + " "
	}

	lines = append(lines, frameEdge(rule, "└", "╯", hintText, width+2, true))

	return strings.Join(lines, "\n")
}

// frameStyles returns the frame and title styles for a dialog.
func (m *model) frameStyles(opts *frameOptions) (rule, title lipgloss.Style) {
	th := m.state.Theme()

	ruleColor, titleColor := th.Info, th.Primary

	switch opts.Style {
	case dialog.StyleInfo:
		ruleColor, titleColor = th.Secondary, th.Info
	case dialog.StyleWarn:
		ruleColor, titleColor = th.Warning, th.Warning
	case dialog.StyleDanger:
		ruleColor, titleColor = th.Error, th.Error
	}

	return lipgloss.NewStyle().Foreground(ruleColor),
		lipgloss.NewStyle().Foreground(titleColor).Bold(true)
}

// frameEdge renders a frame edge with optional text.
func frameEdge(rule lipgloss.Style, start, end, text string, width int, right bool) string {
	inner := width - 2

	if budget := inner - 2; lipgloss.Width(text) > budget {
		if budget <= 0 {
			text = ""
		} else {
			text = ansi.Truncate(text, budget, "…")
		}
	}

	w := lipgloss.Width(text)
	if w == 0 {
		return rule.Render(start + strings.Repeat("─", inner) + end)
	}

	if right {
		return rule.Render(start+strings.Repeat("─", inner-w-2)) +
			text +
			rule.Render(strings.Repeat("─", 2)+end)
	}

	return rule.Render(start+strings.Repeat("─", 2)) +
		text +
		rule.Render(strings.Repeat("─", inner-w-2)+end)
}

// padTo pads s to width cells.
func padTo(s string, width int) string {
	if width <= 0 {
		return ""
	}

	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return ansi.Truncate(s, width, "")
	}

	return s + strings.Repeat(" ", gap)
}
