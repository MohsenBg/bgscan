// Package help renders keyboard help for a component.
// A Model shows a HelpSet in short mode as a few compact rows, or in
// long mode as a scrollable list with wrapped labels. In short mode
// the help key opens the long view in a dialog.
package help

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
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

// HelpSet holds the hints shown in the short and long help views.
type HelpSet struct {
	Short keys.Hints // Short holds the hints shown in the short help view.
	Long  keys.Hints // Long holds the hints shown in the long help view.
}

// HelpMode selects which help view a Model renders.
type HelpMode string

const (
	ShortHelp HelpMode = "short" // ShortHelp shows a compact set of hints in a few rows.
	LongHelp  HelpMode = "long"  // LongHelp shows every hint in a scrollable list.
)

// Model is a help component that renders a HelpSet as short or long help.
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

	vp  viewport.Model
	env *ui.Env

	dirty bool
	cache string
}

// Option configures a Model at construction time.
type Option func(*Model)

// New creates a help component.
func New(env *ui.Env, title string, set HelpSet, opts ...Option) *Model {
	width := max(1, env.Layout.Body().Width)

	m := &Model{
		id:              ui.NewComponentID(),
		title:           title,
		mode:            ShortHelp,
		help:            set,
		env:             env,
		shortHelpWidth:  width,
		shortHelpHeight: defaultShortHelpHeight,
		shortWidthAuto:  true,
		longHelpWidth:   defaultLongHelpWidth,
		longHelpHeight:  defaultLongHelpHeight,
		dirty:           true,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.mode == LongHelp {
		m.initViewport()
	}

	return m
}

// ID returns the component ID of the model.
func (m *Model) ID() ui.ComponentID { return m.id }

// Init returns no initial command.
func (m *Model) Init() tea.Cmd { return nil }

// Mode returns the key mode of the component, which is always normal mode.
func (m *Model) Mode() ui.KeyMode { return ui.KeyModeNormal }

// Name returns a name built from the help mode and the title.
func (m *Model) Name() string {
	return "help/" + string(m.mode) + "/" + m.title
}

// OnClose returns no command when the component closes.
func (m *Model) OnClose() tea.Cmd { return nil }

// WithMode sets the help mode the model starts in.
func WithMode(mode HelpMode) Option {
	return func(m *Model) { m.mode = mode }
}

// WithShortHelpWidth sets a fixed short help width and turns off automatic width sizing.
func WithShortHelpWidth(width int) Option {
	return func(m *Model) {
		m.shortHelpWidth = max(1, width)
		m.shortWidthAuto = false
	}
}

// WithShortHelpHeight sets the maximum number of rows in the short help view.
func WithShortHelpHeight(height int) Option {
	return func(m *Model) { m.shortHelpHeight = max(1, height) }
}

// WithLongHelpWidth sets the preferred width of the long help view.
func WithLongHelpWidth(width int) Option {
	return func(m *Model) { m.longHelpWidth = max(1, width) }
}

// WithLongHelpHeight sets the preferred height of the long help view.
func WithLongHelpHeight(height int) Option {
	return func(m *Model) { m.longHelpHeight = max(1, height) }
}

// rebuild refreshes the long-help viewport from the live theme.
func (m *Model) rebuild() {
	if m.mode == LongHelp {
		m.initViewport()
	}
	m.dirty = true
}

// SetShortWidth sets the short-help width manually.
func (m *Model) SetShortWidth(w int) {
	w = max(1, w)
	m.shortWidthAuto = false

	if m.shortHelpWidth == w {
		return
	}
	m.shortHelpWidth = w
	m.dirty = true
}

func (m *Model) longViewportWidth() int {
	width := m.longHelpWidth - framePaddingHorizontal
	if body := m.env.Layout.Body().Width; body > 0 {
		width = min(width, body)
	}
	return max(1, width)
}

func (m *Model) longViewportHeight() int {
	height := m.longHelpHeight - framePaddingVertical
	if body := m.env.Layout.Body().Height; body > 0 {
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
	m.dirty = true
}

func (m *Model) openDialogFullHelp() tea.Msg {
	return dialog.OpenDialog(
		New(
			m.env,
			m.title,
			m.help,
			WithMode(LongHelp),
			WithLongHelpWidth(m.longHelpWidth),
			WithLongHelpHeight(m.longHelpHeight),
		),
		dialog.WithTitle(m.title+" [Help]"),
		dialog.WithStyle(dialog.StyleInfo),
		dialog.WithOffset(0, -2),
		dialog.WithAnimation(dialog.AnimSplit),
	)
}

func (m *Model) closeDialogFullHelp() tea.Msg {
	return ui.CloseComponentMsg{ID: m.id}
}
