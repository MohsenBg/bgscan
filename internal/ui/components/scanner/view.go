package scanner

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/tabs"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// View renders the scanner UI.
func (m *Model) View() string {
	if m.closing {
		body := m.session.Env().Layout.Body()
		return lipgloss.NewStyle().
			Width(body.Width).
			Height(body.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("⏳ Closing scanner, please wait…")
	}

	th := m.session.Env().Theme

	tab := m.renderTab(&th)
	progressbar := m.renderProgressBar(&th)
	panels := m.renderPanels(&th)

	return lipgloss.NewStyle().
		Align(lipgloss.Center).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			tab,
			"",
			progressbar,
			"",
			panels,
			m.current().table.View(),
		))
}

func (m *Model) renderProgressBar(th *theme.Theme) string {
	stage := m.current()
	p := m.current().progress
	width := m.width

	container := lipgloss.NewStyle().
		Align(lipgloss.Center)

	total := formatCount(p.Total)
	scanned := formatCount(p.Processed)
	left := scanned + styleMuted(th).Render(" / "+total)

	sparkline := styleRate(th).Render(fmt.Sprintf("%.2f IP/s", p.RatePerSec)) + " " +
		styleSparkline(th).Render(stage.sparkline.View())
	eta := styleETA(th).Render("ETA " + formatDuration(p.ETA))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		container.Render(stage.progressBar.View()),
		"",
		threeColumns(left, eta, sparkline, width),
	)
}

func (m *Model) renderTab(th *theme.Theme) string {
	title := m.renderStatusTitle(th)
	if t, ok := m.tabs.(*tabs.Model[int]); ok {
		t.SetTitle(title)
	}
	return m.tabs.View()
}

type statusView struct {
	icon  string
	label string
	style func(th *theme.Theme, pulse int) lipgloss.Style
}

var (
	statusScanning = statusView{"●", "scanning", styleScanning}
	statusPaused   = statusView{"‖", "paused", stylePaused}
	statusError    = statusView{"✗", "error", styleError}
	statusWaiting  = statusView{"○", "waiting", styleWaiting}
	statusDone     = statusView{"✓", "done", styleDone}
)

func (v statusView) render(th *theme.Theme, pulse int) string {
	return v.style(th, pulse).Render(v.icon + " " + v.label)
}

func (m *Model) renderStatusTitle(th *theme.Theme) string {
	if m.scanner.IsPaused() {
		return statusPaused.render(th, m.pulse)
	}
	switch m.current().status {
	case StatusScanning:
		return statusScanning.render(th, m.pulse)
	case StatusError:
		return statusError.render(th, m.pulse)
	case StatusWaiting:
		return statusWaiting.render(th, m.pulse)
	case StatusEnded:
		return statusDone.render(th, m.pulse)
	}
	return ""
}

func threeColumns(left, center, right string, width int) string {
	wl := lipgloss.Width(left)
	wc := lipgloss.Width(center)
	wr := lipgloss.Width(right)

	centerStart := (width - wc) / 2

	gap1 := centerStart - wl
	gap2 := width - centerStart - wc - wr

	if gap1 < 1 || gap2 < 1 {
		free := max(width-wl-wc-wr, 2)
		gap1 = free / 2
		gap2 = free - gap1
	}

	return left + strings.Repeat(" ", gap1) +
		center + strings.Repeat(" ", gap2) +
		right
}
