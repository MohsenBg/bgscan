package scanner

import (
	"math"

	"charm.land/lipgloss/v2"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

func styleScanned(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(th.Yellow)
}

func styleRemaining(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(th.Info)
}

func styleFound(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(th.Success)
}

func styleElapsed(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(th.Purple)
}

func styleScanning(th *theme.Theme, pulse int) lipgloss.Style {
	pulse = max(0, min(pulse, pulseSteps-1))

	colors := lipgloss.Blend1D(pulseSteps, th.Success, th.Muted)

	idx := int(math.Round(float64(pulse) * pulseStrength))
	return lipgloss.NewStyle().Foreground(colors[idx])
}

func stylePaused(th *theme.Theme, pulse int) lipgloss.Style {
	pulse = max(0, min(pulse, pulseSteps-1))

	colors := lipgloss.Blend1D(pulseSteps, th.Warning, th.Muted)

	idx := int(math.Round(float64(pulse) * pulseStrength))
	return lipgloss.NewStyle().Foreground(colors[idx])
}

func styleError(th *theme.Theme, pulse int) lipgloss.Style {
	pulse = max(0, min(pulse, pulseSteps-1))

	colors := lipgloss.Blend1D(pulseSteps, th.Error, th.Muted)

	idx := int(math.Round(float64(pulse) * pulseStrength))
	return lipgloss.NewStyle().Foreground(colors[idx])
}

func styleWaiting(th *theme.Theme, pulse int) lipgloss.Style {
	pulse = max(0, min(pulse, pulseSteps-1))

	colors := lipgloss.Blend1D(pulseSteps, th.Info, th.Muted)

	idx := int(math.Round(float64(pulse) * pulseStrength))
	return lipgloss.NewStyle().Foreground(colors[idx])
}

func styleDone(th *theme.Theme, pulse int) lipgloss.Style {
	pulse = max(0, min(pulse, pulseSteps-1))

	colors := lipgloss.Blend1D(pulseSteps, th.Primary, th.Muted)

	idx := int(math.Round(float64(pulse) * pulseStrength))
	return lipgloss.NewStyle().Foreground(colors[idx])
}

func styleRate(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Info)
}

func styleSparkline(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Primary)
}

func styleETA(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Orange)
}

func styleMuted(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Muted)
}

func stylePanelTitle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Text)
}

// stylePanelLabel colors a row label inside a panel.
func stylePanelLabel(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Secondary)
}

// styleErrorTotal colors the failed count in the Errors panel title.
func styleErrorTotal(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Error)
}

func hitRateBarStyle(th *theme.Theme, rate float64) lipgloss.Style {
	switch {
	case rate >= 0.80:
		return lipgloss.NewStyle().Foreground(th.Success)
	case rate >= 0.40:
		return lipgloss.NewStyle().Foreground(th.Yellow)
	default:
		return lipgloss.NewStyle().Foreground(th.Error)
	}
}
