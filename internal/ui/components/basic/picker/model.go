package picker

import (
	"os"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
)

// Model is a file picker overlay wrapping BubbleTea's filepicker.Model and
// integrating it with the component/overlay system.
type Model struct {
	id   ui.ComponentID
	name string

	// Overlay title displayed by the layout
	Title string

	// Layout manager used for sizing calculations
	Layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set

	// Underlying BubbleTea file picker
	FilePicker filepicker.Model

	// Callback triggered when a file is selected
	OnSelect OnSelect
}

// Init initializes the underlying file picker component.
func (m *Model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

// CloseCmd emits CloseComponentMsg; the router removes the overlay.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// New creates a file picker overlay. baseDir defaults to the user's home
// directory when empty; allowType restricts selectable extensions; onSelect
// runs after a file is selected (a no-op is used if nil).
func New(deps ui.Deps, title string, baseDir string, allowType []string, onSelect OnSelect) *Model {
	p := filepicker.New()

	if baseDir != "" {
		p.CurrentDirectory = baseDir
	} else {
		p.CurrentDirectory, _ = os.UserHomeDir()
	}

	if len(allowType) > 0 {
		p.AllowedTypes = allowType
	}

	// Ensure callback is never nil
	if onSelect == nil {
		onSelect = func(path string) tea.Cmd { return nil }
	}

	p.ShowPermissions = true
	p.AutoHeight = false
	p.SetHeight(pickerHeight(deps.Layout))

	return &Model{
		id:         ui.NewComponentID(),
		name:       "Pick File",
		Title:      title,
		Layout:     deps.Layout,
		theme:      deps.Theme,
		log:        deps.Log,
		FilePicker: p,
		OnSelect:   onSelect,
	}
}

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) { m.theme = th }

func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns NormalMode for standard keyboard navigation.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
