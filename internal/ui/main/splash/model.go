package splash

import (
	"image/color"
	"strings"
	"time"

	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

var (
	glitchChars = []rune("▓▒░╬╪┼╫┃━╳▄▀")
	glitchSet   = string(glitchChars)
)

var logoArt = strings.Split(strings.TrimSpace(`
██████╗  ██████╗      ███████╗ ██████╗ █████╗ ███╗   ██╗
██╔══██╗██╔════╝      ██╔════╝██╔════╝██╔══██╗████╗  ██║
██████╔╝██║  ███╗     ███████╗██║     ███████║██╔██╗ ██║
██╔══██╗██║   ██║     ╚════██║██║     ██╔══██║██║╚██╗██║
██████╔╝╚██████╔╝     ███████║╚██████╗██║  ██║██║ ╚████║
╚═════╝  ╚═════╝      ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`), "\n")

var logoWidth = func() int {
	w := 0
	for _, l := range logoArt {
		if n := len([]rune(l)); n > w {
			w = n
		}
	}
	return w
}()

const (
	tickInterval = 30 * time.Millisecond

	animFrames  = 80
	holdFrames  = 150
	totalFrames = animFrames + holdFrames
)

const (
	sweepStartProgress = 0.72
	sweepSpan          = 1 - sweepStartProgress
	sweepOvershoot     = 6

	corruptionChance  = 0.5
	idleFlickerChance = 0.001

	shiftChanceAnimating = 0.65
	shiftChanceHolding   = 0.03

	shakeCutoff       = 0.05
	artifactRowChance = 0.9
)

type tickMsg time.Time

type model struct {
	id    ui.ComponentID
	name  string
	state *ui.AppState
	frame int
}

func (m model) ID() ui.ComponentID { return m.id }
func (m model) Mode() env.Mode     { return env.NormalMode }
func (m model) Name() string       { return m.name }
func (m model) OnClose() tea.Cmd   { return nil }

func New(state *ui.AppState) ui.Component {
	return &model{
		id:    ui.NewComponentID(),
		name:  "splash screen",
		state: state,
		frame: 0,
	}
}

func (m model) Init() tea.Cmd { return tickCmd() }

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func baseColor() color.Color   { return theme.Current().Success }
func glitchColor() color.Color { return theme.Current().Muted }
func accentColor() color.Color { return theme.Current().Secondary }
