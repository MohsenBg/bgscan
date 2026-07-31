package scanner

import (
	"fmt"
	"time"

	"bgscan/internal/core/scanner/engine"

	"charm.land/lipgloss/v2"
)

// View renders the scanner UI.
func (m *Model) View() string {
	idx := m.currentTab
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.tabs.View(),
		m.renderProgress(idx),
		m.ipViewers[idx].View(),
	)
}

func (m *Model) renderProgress(idx int) string {
	p := m.progressInfo[idx]
	width := m.state.Layout.Body.Width

	stats := m.renderStatsRow(p)
	status := m.renderStatusRow()

	container := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		container.Render(stats),
		container.Padding(1, 0).Render(status),
		container.Render(m.progress[idx].View()),
	)
}

func (m *Model) renderStatsRow(p engine.Progress) string {
	left := uint64(0)
	if p.Total > p.Processed {
		left = p.Total - p.Processed
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		scannedStyle().Render(fmt.Sprintf("scanned: %s", formatCount(p.Processed))),
		separatorStyle().Render(" | "),
		leftStyle().Render(fmt.Sprintf("left: %s", formatCount(left))),
		separatorStyle().Render(" | "),
		foundStyle().Render(fmt.Sprintf("found: %s", formatCount(p.Succeed))),
		separatorStyle().Render(" | "),
		elapsedStyle().Render(fmt.Sprintf("elapsed: %s", formatDuration(p.Elapsed))),
	)
}

func (m *Model) renderStatusRow() string {
	return elapsedEndStyle().Render(m.statusText())
}

func (m *Model) statusText() string {
	idx := m.currentTab
	status := m.status[idx]
	p := m.progressInfo[idx]

	switch status {
	case StatusPreProcess:
		return "preparing scan..."
	case StatusScanning:
		if m.scn.IsPaused() {
			return "scan paused..."
		}
		return m.estimateRemaining(p)
	case StatusEnded:
		return "scan completed"
	case StatusError:
		if m.scanError != nil {
			return fmt.Sprintf("scan error: %v", m.scanError)
		}
		return "scan error"
	default:
		return "starting scan..."
	}
}

func (m *Model) estimateRemaining(p engine.Progress) string {
	if p.ETA <= 0 || p.RatePerSec <= 0 {
		return "estimating remaining time..."
	}

	rateStr := fmt.Sprintf("[%.2fIP/S]", p.RatePerSec)
	rateStr = leftStyle().Render(rateStr)
	return fmt.Sprintf("estimated remaining: %s %s", formatDuration(p.ETA), rateStr)
}

// formatCount formats a uint64 as a short human-readable string.
// Supports up to Exa (10^18), which gracefully handles 2^64 (~18.45E).
func formatCount(n uint64) string {
	switch {
	case n < 10_000:
		return fmt.Sprintf("%d", n)
	case n < 1_000_000:
		return fmt.Sprintf("%.2fK", float64(n)/1_000)
	case n < 1_000_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n < 1_000_000_000_000:
		return fmt.Sprintf("%.2fB", float64(n)/1_000_000_000)
	case n < 1_000_000_000_000_000:
		return fmt.Sprintf("%.2fT", float64(n)/1_000_000_000_000)
	case n < 1_000_000_000_000_000_000:
		return fmt.Sprintf("%.2fP", float64(n)/1_000_000_000_000_000) // Peta
	default:
		return fmt.Sprintf("%.2fE", float64(n)/1_000_000_000_000_000_000) // Exa
	}
}

// formatDuration formats a duration into a human-readable string that scales
// from seconds all the way up to years, dropping lower units to prevent UI clutter.
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	const (
		minute = time.Minute
		hour   = time.Hour
		day    = 24 * time.Hour
		year   = 365*day + 6*time.Hour
	)

	switch {
	case d < minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < hour:
		m := d / minute
		s := (d % minute) / time.Second
		return fmt.Sprintf("%dm %ds", m, s)
	case d < day:
		h := d / hour
		m := (d % hour) / minute
		s := (d % minute) / time.Second
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	case d < year:
		days := d / day
		h := (d % day) / hour
		m := (d % hour) / minute
		return fmt.Sprintf("%dd %dh %dm", days, h, m)
	default:
		years := d / year
		days := (d % year) / day
		h := (d % day) / hour
		return fmt.Sprintf("%dy %dd %dh", years, days, h)
	}
}
