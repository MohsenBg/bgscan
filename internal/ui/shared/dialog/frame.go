package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

const (
	// minFrameWidth is the narrowest frame the drawer will emit.
	minFrameWidth = 16

	// hintSeparator joins the hints drawn on the frame's bottom border.
	hintSeparator = "   "
)

// FrameOptions is the chrome drawn around a dialog's content.
type FrameOptions struct {
	// Title is embedded in the top border, left aligned.
	Title string

	// Style picks the border and title palette.
	Style DialogStyle

	// Hints are drawn on the bottom border, right aligned.
	Hints []keys.Hint

	// Width is the space available to the content. The frame never grows
	// past it, so a dialog always fits the body.
	Width int
}

// RenderFrame draws the frame, the title and the hints around content.
//
// The rendering is intentionally free of animation state: the animation is
// applied later by Composite, when the frame is placed over the screen.
func RenderFrame(content string, o FrameOptions, th *theme.Theme) string {
	innerWidth := max(minFrameWidth, o.Width)

	body := lipgloss.NewStyle().
		MaxWidth(innerWidth).
		Padding(1).
		Render(content)

	width := max(lipgloss.Width(body), minFrameWidth)

	rule, titleStyle := frameStyles(o.Style, th)

	title := ""
	if o.Title != "" {
		title = titleStyle.Render(" " + o.Title + " ")
	}

	rows := strings.Split(body, "\n")
	lines := make([]string, 0, len(rows)+2)

	lines = append(lines, frameEdge(rule, "╭", "┐", title, width+2, false))

	for _, row := range rows {
		lines = append(lines, rule.Render("│")+padTo(row, width)+rule.Render("│"))
	}

	hintText := ""
	if hint := HintText(o.Hints, th); hint != "" {
		hintText = " " + hint + " "
	}

	lines = append(lines, frameEdge(rule, "└", "╯", hintText, width+2, true))

	return strings.Join(lines, "\n")
}

// HintText renders key hints for the frame's bottom border.
func HintText(hints []keys.Hint, th *theme.Theme) string {
	parts := make([]string, 0, len(hints))

	for _, h := range hints {
		parts = append(parts, Hint(h, th))
	}

	return strings.Join(parts, hintSeparator)
}

// Hint renders a key and its label.
func Hint(h keys.Hint, th *theme.Theme) string {
	return lipgloss.NewStyle().
		Foreground(th.Primary).
		Bold(true).
		Render(h.Key) +
		" " +
		lipgloss.NewStyle().
			Foreground(th.Muted).
			Render(h.Label)
}

// frameStyles returns the frame and title styles for a dialog.
func frameStyles(s DialogStyle, th *theme.Theme) (rule, title lipgloss.Style) {
	ruleColor, titleColor := th.Info, th.Primary

	switch s {
	case StyleInfo:
		ruleColor, titleColor = th.Secondary, th.Info
	case StyleWarn:
		ruleColor, titleColor = th.Warning, th.Warning
	case StyleDanger:
		ruleColor, titleColor = th.Error, th.Error
	case StyleSuccess:
		ruleColor, titleColor = th.Success, th.Success
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
