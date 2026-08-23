package splash

import (
	"math"
	"math/rand"
	"strings"

	"bgscan/internal/core/config"

	lipgloss "charm.land/lipgloss/v2"
)

func (m model) View() string {
	progress, intensity := 1.0, 0.0
	holding := m.frame >= animFrames
	if !holding {
		progress = float64(m.frame) / animFrames
		intensity = math.Pow(1-progress, 1.6)
	}

	lines := make([]string, len(logoArt))
	for li, line := range logoArt {
		lines[li] = renderLine(line, progress, intensity, holding)
	}

	if intensity > shakeCutoff {
		pad := strings.Repeat(" ", rand.Intn(int(intensity*5)+1))
		for i := range lines {
			lines[i] = pad + lines[i]
		}
	}

	var bottom string
	if holding {
		bottom = "\n" + versionLine(m.frame-animFrames)
	} else {
		bottom = artifactRow(intensity)
	}

	content := strings.Join([]string{
		artifactRow(intensity),
		strings.Join(lines, "\n"),
		bottom,
	}, "\n")

	return lipgloss.Place(
		m.state.Layout.Terminal.Width,
		m.state.Layout.Terminal.Height,
		lipgloss.Center,
		lipgloss.Center, content,
	)
}

func renderLine(line string, progress, intensity float64, holding bool) string {
	runes := []rune(line)

	corrChance := intensity * corruptionChance
	if holding {
		corrChance = idleFlickerChance
	}
	for i, r := range runes {
		if r != ' ' && rand.Float64() < corrChance {
			runes[i] = glitchChars[rand.Intn(len(glitchChars))]
		}
	}

	s := string(runes)
	switch {
	case holding && rand.Float64() < shiftChanceHolding:
		s = shiftLine(s, rand.Intn(5)-2)
	case !holding && rand.Float64() < intensity*shiftChanceAnimating:
		shift := rand.Intn(int(intensity*9)+1) - int(intensity*4)
		s = shiftLine(s, shift)
	}

	return colorize(s, progress, holding)
}

func colorize(s string, progress float64, holding bool) string {
	sweepActive := progress > sweepStartProgress
	sweepPos := -10
	if sweepActive {
		sweepFrac := (progress - sweepStartProgress) / sweepSpan
		sweepPos = int(sweepFrac*float64(logoWidth+2*sweepOvershoot)) - sweepOvershoot
	}

	runes := []rune(s)
	var sb strings.Builder
	for col, ch := range runes {
		if ch == ' ' {
			sb.WriteByte(' ')
			continue
		}

		var st lipgloss.Style
		switch {
		case col >= sweepPos-1 && col <= sweepPos+1:
			st = lipgloss.NewStyle().Bold(true).Foreground(accentColor())
		case strings.ContainsRune(glitchSet, ch):
			if holding {
				st = lipgloss.NewStyle().Foreground(accentColor())
			} else {
				st = lipgloss.NewStyle().Foreground(glitchColor())
			}
		default:
			st = lipgloss.NewStyle().Foreground(baseColor())
		}
		sb.WriteString(st.Render(string(ch)))
	}
	return sb.String()
}

func artifactRow(intensity float64) string {
	if intensity <= 0 || rand.Float64() > intensity*artifactRowChance {
		return " "
	}
	st := lipgloss.NewStyle().Foreground(glitchColor())
	var sb strings.Builder
	for i := 0; i < 1+rand.Intn(12); i++ {
		sb.WriteString(st.Render(string(glitchChars[rand.Intn(len(glitchChars))])))
	}
	return strings.Repeat(" ", rand.Intn(20)) + sb.String()
}

func versionLine(holdFrame int) string {
	var color string
	switch {
	case holdFrame < 7:
		color = "#232833"
	case holdFrame < 14:
		color = "#4c566a"
	default:
		color = "#7b8794"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(config.AppVersion)
}

func shiftLine(s string, shift int) string {
	switch {
	case shift > 0:
		return strings.Repeat(" ", shift) + s
	case shift < 0:
		r := []rune(s)
		if -shift >= len(r) {
			return ""
		}
		return string(r[-shift:])
	default:
		return s
	}
}
