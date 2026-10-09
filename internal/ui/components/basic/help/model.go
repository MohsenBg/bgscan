// Package help renders short and long keyboard help panels.
package help

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

const (
	defaultShortHelpHeight = 2
	defaultLongHelpWidth   = 50
	defaultLongHelpHeight  = 30

	framePaddingVertical   = 1
	framePaddingHorizontal = 2
	keyLabelGap            = 5
	shortHelpPadding       = 1
)

type HelpSet struct {
	Short keys.Hints
	Long  keys.Hints
}

type HelpMode string

const (
	ShortHelp HelpMode = "short"
	LongHelp  HelpMode = "long"
)

type Model struct {
	id    ui.ComponentID
	title string
	mode  HelpMode
	help  HelpSet

	shortHelpWidth  int
	shortHelpHeight int
	shortWidthAuto  bool

	longHelpWidth  int
	longHelpHeight int

	vp   viewport.Model
	deps ui.Deps
}

type Option func(*Model)

// New creates a help component.
func New(deps ui.Deps, title string, set HelpSet, opts ...Option) *Model {
	// Layout is optional: a caller that has not been sized yet still gets
	// a usable component, and the first resize corrects the width.
	width := 1
	if deps.Layout != nil {
		width = max(1, deps.Layout.BodyContentWidth())
	}

	m := &Model{
		id:              ui.NewComponentID(),
		title:           title,
		mode:            ShortHelp,
		help:            set,
		deps:            deps,
		shortHelpWidth:  width,
		shortHelpHeight: defaultShortHelpHeight,
		shortWidthAuto:  true,
		longHelpWidth:   defaultLongHelpWidth,
		longHelpHeight:  defaultLongHelpHeight,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.mode == LongHelp {
		m.initViewport()
	}

	return m
}

func (m *Model) ID() ui.ComponentID { return m.id }

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Mode() env.Mode { return env.NormalMode }

func (m *Model) Name() string {
	return "help/" + string(m.mode) + "/" + m.title
}

func (m *Model) OnClose() tea.Cmd { return nil }

func WithMode(mode HelpMode) Option {
	return func(m *Model) { m.mode = mode }
}

func WithShortHelpWidth(width int) Option {
	return func(m *Model) {
		m.shortHelpWidth = max(1, width)
		m.shortWidthAuto = false
	}
}

func WithShortHelpHeight(height int) Option {
	return func(m *Model) { m.shortHelpHeight = max(1, height) }
}

func WithLongHelpWidth(width int) Option {
	return func(m *Model) { m.longHelpWidth = max(1, width) }
}

func WithLongHelpHeight(height int) Option {
	return func(m *Model) { m.longHelpHeight = max(1, height) }
}

// SetTheme updates the theme and rebuilds long help when needed.
func (m *Model) SetTheme(t *theme.Theme) {
	m.deps.Theme = t

	if m.mode == LongHelp {
		m.initViewport()
	}
}

// SetShortWidth sets the short-help width manually.
func (m *Model) SetShortWidth(w int) {
	m.shortHelpWidth = max(1, w)
	m.shortWidthAuto = false
}

func (m *Model) longViewportWidth() int {
	width := m.longHelpWidth - framePaddingHorizontal
	if body := m.deps.Layout.BodyContentWidth(); body > 0 {
		width = min(width, body)
	}
	return max(1, width)
}

func (m *Model) longViewportHeight() int {
	height := m.longHelpHeight - framePaddingVertical
	if body := m.deps.Layout.BodyContentHeight(); body > 0 {
		height = min(height, body)
	}
	return max(1, height)
}

func (m *Model) longKeyWidth() int {
	w := 0
	for _, hint := range m.help.Long {
		w = max(w, lipgloss.Width(hint.Key))
	}
	return w
}

func (m *Model) longNaturalWidth() int {
	labelWidth := lipgloss.Width(m.title)
	for _, hint := range m.help.Long {
		labelWidth = max(labelWidth, lipgloss.Width(hint.Label))
	}
	return max(1, m.longKeyWidth()+keyLabelGap+labelWidth)
}

type longLayout struct {
	width   int
	height  int
	content string
}

func (m *Model) longLayout() longLayout {
	width := max(1, min(m.longViewportWidth(), m.longNaturalWidth()))
	lines := m.longLines(width)

	height := max(1, min(len(lines), m.longViewportHeight()))

	return longLayout{
		width:   width,
		height:  height,
		content: strings.Join(lines, "\n"),
	}
}

func (m *Model) initViewport() {
	layout := m.longLayout()

	m.vp = viewport.New(
		viewport.WithWidth(layout.width),
		viewport.WithHeight(layout.height),
	)
	m.vp.SetContent(layout.content)
}

func (m *Model) updateViewportSize() {
	layout := m.longLayout()

	m.vp.SetWidth(layout.width)
	m.vp.SetHeight(layout.height)
	m.vp.SetContent(layout.content)
}

func (m *Model) openDialogFullHelp() tea.Msg {
	return dialog.OpenDialog(
		New(
			m.deps,
			m.title,
			m.help,
			WithMode(LongHelp),
			WithLongHelpWidth(m.longHelpWidth),
			WithLongHelpHeight(m.longHelpHeight),
		),
		dialog.WithTitle(m.title+" [Help]"),
		dialog.WithStyle(dialog.StyleInfo),
		dialog.WithOffset(0, -2),
	)
}

func (m *Model) closeDialogFullHelp() tea.Msg {
	return ui.CloseComponentMsg{ID: m.id}
}
