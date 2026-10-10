package table

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func (m *Model) tableStyles() table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.theme.Muted).
		Foreground(m.theme.Info).
		BorderBottom(true).
		Padding(0, 1, 0, cellPadLeft)

	s.Cell = s.Cell.Padding(0, 1, 0, cellPadLeft).Bold(false)

	s.Selected = s.Selected.
		Foreground(m.theme.Text).
		Background(m.theme.Selected).
		Height(1).
		Bold(true)

	return s
}

func (m *Model) titleStyles(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(m.theme.Info).
		Bold(true).
		Padding(1, 0)
}

// filterStyle is given the full table width so every block handed to
// JoinVertical has the same width. With equal-width blocks, the join becomes
// a no-op and the outer Align(Center) is exact.
func (m *Model) filterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(m.tableWidthLocked()).
		Align(lipgloss.Left).
		Foreground(m.theme.Info)
}

func (m *Model) tableViewStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		BorderForeground(m.theme.Info)
}

// frameH reports the horizontal cells consumed by the table frame.
// It derives the value from tableViewStyle so width calculations stay in sync
// with the rendered chrome.
func (m *Model) frameH() int {
	st := m.tableViewStyle()
	return st.GetHorizontalBorderSize() + st.GetHorizontalPadding()
}

// frameV reports the vertical cells consumed by the table frame.
func (m *Model) frameV() int {
	st := m.tableViewStyle()
	return st.GetVerticalBorderSize() + st.GetVerticalPadding()
}

func (m *Model) helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(m.tableWidthLocked()).
		Height(helpHeight).
		Align(lipgloss.Center, lipgloss.Bottom)
}

// mixColor blends fg over bg. Terminals have no alpha channel, so reduced
// opacity is approximated by mixing the two colors.
func mixColor(fg, bg color.Color, opacity float64) color.Color {
	opacity = max(0, min(1, opacity))

	fr, fgGreen, fb, _ := fg.RGBA()
	br, bgGreen, bb, _ := bg.RGBA()

	blend := func(f, b uint32) uint8 {
		return uint8(float64(f>>8)*opacity + float64(b>>8)*(1-opacity) + 0.5)
	}

	return color.RGBA{
		R: blend(fr, br),
		G: blend(fgGreen, bgGreen),
		B: blend(fb, bb),
		A: 255,
	}
}

func (m *Model) applyStylesLocked() {
	s := m.tableStyles()
	m.BubbleTable.SetStyles(s)

	const sentinel = "§"
	m.selPrefix, _, _ = strings.Cut(s.Selected.Render(sentinel), sentinel)

	m.marker = lipgloss.NewStyle().
		Foreground(mixColor(m.theme.Primary, m.theme.Selected, markerOpacity)).
		Background(m.theme.Selected).
		Bold(true).
		Render(markerGlyph)
}

// markSelected draws the bar over the left padding of the selected row.
// It edits only the rendered string, so row data is not copied or mutated.
func (m *Model) markSelected(view string) string {
	if m.selPrefix == "" {
		return view
	}

	needle := m.selPrefix + strings.Repeat(" ", cellPadLeft)
	before, after, ok := strings.Cut(view, needle)
	if !ok {
		return view
	}

	return before + m.marker + m.selPrefix + after
}
