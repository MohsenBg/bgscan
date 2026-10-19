package sparkline

import (
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// gradientStyles returns one style per cell, fading from ProgressStart on the
// oldest (left) cell to ProgressEnd on the newest (right) cell.
//
// The color depends on the cell's position on screen, not on the sample, so it
// can never flicker while the chart scrolls.
func gradientStyles(th *theme.Theme, n int) []lipgloss.Style {
	styles := make([]lipgloss.Style, n)

	colors := lipgloss.Blend1D(n, th.Primary, th.Info)
	for i := range styles {
		style := lipgloss.NewStyle()
		if len(colors) > 0 {
			style = style.Foreground(colors[min(i, len(colors)-1)])
		}
		styles[i] = style
	}

	return styles
}
