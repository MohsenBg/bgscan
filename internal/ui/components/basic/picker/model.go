package picker

import (
	"os"

	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

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

	// Underlying BubbleTea file picker
	FilePicker filepicker.Model

	// Callback triggered when a file is selected
	OnSelect OnSelect
}

// Init initializes the underlying file picker component.
func (m *Model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the human‑readable component name.
func (m *Model) Name() string {
	return m.name
}

// CloseCmd returns a command that closes this overlay component.
//
// The command emits a `ui.CloseComponentMsg` which the application
// router handles to remove the overlay from the stack.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// New creates a file picker overlay. baseDir defaults to the user's home
// directory when empty; allowType restricts selectable extensions; onSelect
// runs after a file is selected (a no-op is used if nil).
func New(layout *layout.Layout, title string, baseDir string, allowType []string, onSelect OnSelect) *Model {
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
	p.SetHeight(pickerHeight(layout))

	return &Model{
		id:         ui.NewComponentID(),
		name:       "Pick File",
		Title:      title,
		Layout:     layout,
		FilePicker: p,
		OnSelect:   onSelect,
	}
}

// OnClose is called when the overlay is removed from the stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the input mode required by this component.
// The file picker operates in NormalMode, allowing standard
// keyboard navigation through files and directories.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
