package scanner

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/core/scanner/engine"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

const (
	panelGap = 2

	// scanPanelRatio is the share of the row width given to the Scan panel.
	scanPanelRatio = 40

	// frameBorderW is the width of a frame's left and right border columns.
	frameBorderW = 2
	// panelPadH is the horizontal padding inside a panel.
	panelPadH = 1

	// errCountWidth is the width of the count column in the Errors panel.
	errCountWidth = 6
	// errPctWidth is the width of the percentage column in the Errors panel.
	errPctWidth = 4

	filledBlock = "━"
	emptyBlock  = "−"
)

// errKind is one failure mode shown in the Errors panel.
type errKind struct {
	name  string
	count uint64
	color color.Color
}

// renderPanels renders the Scan and Errors summary panels side by side.
func (m *Model) renderPanels(th *theme.Theme) string {
	p := m.current().progress

	scanWidth := m.width * scanPanelRatio / 100
	errWidth := m.width - scanWidth - panelGap

	scanBody := scanPanelBody(p, th, contentWidth(scanWidth))
	errBody := errorsPanelBody(p, th, contentWidth(errWidth))

	// Pad both bodies to the same height so the borders line up.
	scanBody, errBody = equalizeHeight(scanBody, errBody)

	scan := m.renderFrame(scanBody, th, frameSpec{
		left:  stylePanelTitle(th).Render(" Scan "),
		width: scanWidth,
		padH:  panelPadH,
	})
	errs := m.renderFrame(errBody, th, frameSpec{
		left:  stylePanelTitle(th).Render(" Errors "),
		right: styleErrorTotal(th).Render(" " + formatCount(p.Errors.Total()) + " failed "),
		width: errWidth,
		padH:  panelPadH,
	})

	return lipgloss.JoinHorizontal(lipgloss.Top, scan, strings.Repeat(" ", panelGap), errs)
}

// contentWidth returns the usable width inside a frame of the given outer width.
func contentWidth(outer int) int {
	return outer - frameBorderW - 2*panelPadH
}

// equalizeHeight pads the shorter body with blank rows so both render to the
// same height.
func equalizeHeight(a, b string) (string, string) {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	for len(aLines) < len(bLines) {
		aLines = append(aLines, "")
	}
	for len(bLines) < len(aLines) {
		bLines = append(bLines, "")
	}

	return strings.Join(aLines, "\n"), strings.Join(bLines, "\n")
}

// scanPanelBody renders the Scan panel: the progress counters and a bar
// showing the hit rate.
func scanPanelBody(p engine.Progress, th *theme.Theme, width int) string {
	rows := []struct {
		name  string
		value string
		style lipgloss.Style
	}{
		{"scanned", formatCount(p.Processed), styleScanned(th)},
		{"left", formatCount(remaining(p)), styleRemaining(th)},
		{"found", formatCount(p.Succeed), styleFound(th)},
		{"elapsed", formatDuration(p.Elapsed), styleElapsed(th)},
	}

	lines := make([]string, 0, len(rows)+1)
	for _, r := range rows {
		lines = append(lines, statRow(r.name, r.value, width, r.style))
	}
	lines = append(lines, "")
	lines = append(lines,
		stylePanelLabel(th).Render(padRight("hit rate", width-lipgloss.Width(formatHitRate(p))))+
			hitRateBarStyle(th, hitRate(p)).Render(formatHitRate(p)),
	)
	lines = append(lines, renderBar(hitRate(p), width, hitRateBarStyle(th, hitRate(p)), styleMuted(th)))

	return strings.Join(lines, "\n")
}

// statRow renders a label on the left and an already-styled value flush right.
func statRow(name, value string, width int, style lipgloss.Style) string {
	return style.Render(padRight(name, width-lipgloss.Width(value)) + value)
}

// errorsPanelBody renders the Errors panel: one row per failure mode with a
// bar showing its share of the total failures.
func errorsPanelBody(p engine.Progress, th *theme.Theme, width int) string {
	total := p.Errors.Total()

	kinds := []errKind{
		{"timeout", p.Errors.Timeout, th.Error},
		{"refused", p.Errors.Refused, th.Orange},
		{"unreachable", p.Errors.Unreachable, th.Yellow},
		{"closed", p.Errors.Closed, th.Primary},
		{"badresponse", p.Errors.BadResponse, th.Secondary},
		{"other", p.Errors.Other, th.Muted},
	}

	// Row layout: name | count | space | bar | space | pct
	leftW := max(width/2, 20)
	rightW := width - leftW
	nameW := leftW - errCountWidth - 1
	barW := rightW - errPctWidth - 1

	label := stylePanelLabel(th)
	track := styleMuted(th)

	lines := make([]string, 0, len(kinds))
	for _, k := range kinds {
		style := lipgloss.NewStyle().Foreground(k.color)
		pct := errorPercent(k.count, total)

		lines = append(lines,
			style.Render(padRight(k.name, nameW))+
				label.Render(padLeft(formatCount(k.count), errCountWidth))+" "+
				renderBar(pct/100, barW, style, track)+" "+
				style.Render(padLeft(formatErrorPercent(pct), errPctWidth)))
	}

	return strings.Join(lines, "\n")
}

// renderBar renders a horizontal bar of the given width, filled in proportion
// to fraction (clamped to 0..1).
func renderBar(fraction float64, width int, fill, track lipgloss.Style) string {
	if width <= 0 {
		return ""
	}

	fraction = max(0, min(fraction, 1))
	filled := int(math.Round(fraction * float64(width)))

	return fill.Render(strings.Repeat(filledBlock, filled)) +
		track.Render(strings.Repeat(emptyBlock, width-filled))
}

// padRight pads s with spaces on the right to the given width.
func padRight(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// padLeft pads s with spaces on the left to the given width.
func padLeft(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return strings.Repeat(" ", gap) + s
	}
	return s
}

// remaining returns the number of targets left to process.
func remaining(p engine.Progress) uint64 {
	if p.Total > p.Processed {
		return p.Total - p.Processed
	}
	return 0
}

// hitRate returns the share of processed targets that succeeded, as a
// fraction between zero and one.
func hitRate(p engine.Progress) float64 {
	if p.Processed == 0 {
		return 0
	}
	return float64(p.Succeed) / float64(p.Processed)
}

// formatHitRate renders the hit rate as a percentage.
func formatHitRate(p engine.Progress) string {
	return fmt.Sprintf("%.1f%%", hitRate(p)*100)
}

// errorPercent returns the share of total failures a failure mode accounts
// for, as a percentage.
func errorPercent(count, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}

// formatErrorPercent renders a failure share, keeping the smallest shares
// visible.
func formatErrorPercent(pct float64) string {
	switch {
	case pct <= 0:
		return "-"
	case pct < 1:
		return "<1%"
	default:
		return fmt.Sprintf("%.0f%%", pct)
	}
}
