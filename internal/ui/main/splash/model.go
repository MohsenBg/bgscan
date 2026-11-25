// Package splash renders the animated logo shown when the app starts.
// The logo plays a glitch animation, then holds with the app version
// below it. The screen ends after a fixed number of frames, or at once
// when the user presses any key, and reports that with SplashDoneMsg.
package splash

import (
	"image/color"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

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

	// Frame counts for the two phases. The animation runs for animFrames
	// ticks, then the finished logo holds for holdFrames ticks.
	animFrames  = 80
	holdFrames  = 150
	totalFrames = animFrames + holdFrames
)

// Glitch tuning values. Chances are probabilities per rune or per line,
// checked on every frame.
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
	id      ui.ComponentID
	name    string
	session *ui.Session
	frame   int
}

func (m model) ID() ui.ComponentID { return m.id }
func (m model) Mode() ui.KeyMode   { return ui.KeyModeNormal }
func (m model) Name() string       { return m.name }
func (m model) OnClose() tea.Cmd   { return nil }

// New creates the splash screen component.
func New(session *ui.Session) ui.Component {
	return &model{
		id:      ui.NewComponentID(),
		name:    "splash screen",
		session: session,
		frame:   0,
	}
}

func (m model) Init() tea.Cmd { return tickCmd() }

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) baseColor() color.Color   { return m.session.Env().Theme.Success }
func (m model) glitchColor() color.Color { return m.session.Env().Theme.Muted }
func (m model) accentColor() color.Color { return m.session.Env().Theme.Secondary }
