package scanner

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

const (
	// lead is the number of border cells drawn before the left title.
	lead = 2
	// minFill is the smallest number of border cells kept between the left
	minFill = 1
)

// frameSpec describes a single framed panel.
type frameSpec struct {
	left  string
	right string
	width int
	padV  int
	padH  int
}

// renderFrame wraps content in a rounded border with a title left and right
func (m *Model) renderFrame(content string, th *theme.Theme, spec frameSpec) string {
	body := lipgloss.NewStyle().Padding(spec.padV, spec.padH).Render(content)

	inner := lipgloss.Width(body)
	if spec.width > 0 {
		inner = max(spec.width-2, 1)
		body = lipgloss.NewStyle().Padding(spec.padV, spec.padH).Width(inner).Render(content)
	}

	border := lipgloss.NewStyle().Foreground(th.Border)
	side := border.Render("│")

	rows := strings.Split(body, "\n")
	lines := make([]string, 0, len(rows)+2)

	lines = append(lines, topEdge(border, spec.left, spec.right, inner))
	for _, row := range rows {
		lines = append(lines, side+row+side)
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", inner)+"╯"))

	return strings.Join(lines, "\n")
}

// topEdge renders the top border with a title on each end
func topEdge(border lipgloss.Style, left, right string, inner int) string {
	leadLen := min(lead, inner)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	if leadLen+leftWidth+rightWidth+minFill > inner {
		right = ""
		rightWidth = 0
	}

	budget := inner - leadLen - rightWidth - minFill
	switch {
	case budget <= 0:
		left = ""
		leftWidth = 0
	case leftWidth > budget:
		left = ansi.Truncate(left, budget, "…")
		leftWidth = lipgloss.Width(left)
	}

	fill := inner - leadLen - leftWidth - rightWidth
	if fill < 0 {
		fill = 0
	}

	return border.Render("╭"+strings.Repeat("─", leadLen)) +
		left +
		border.Render(strings.Repeat("─", fill)) +
		right +
		border.Render("╮")
}
