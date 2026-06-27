package picker

import (
	"os"

	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

// Model implements a file picker overlay component.
//
// It wraps the BubbleTea `filepicker.Model` and integrates it with the
// application's component system and overlay stack.
//
// Responsibilities:
//   - Display a navigable file picker UI
//   - Restrict selectable file types
//   - Invoke a callback when a file is selected
//   - Close itself through the component manager
type Model struct {
	// Component metadata
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

// New creates a new file picker overlay.
//
// Parameters:
//
//	layout    — UI layout manager used to compute component sizing
//	title     — overlay title displayed in the UI
//	baseDir   — initial directory to open (defaults to user home)
//	allowType — allowed file extensions (e.g. []string{".txt",".csv"})
//	onSelect  — callback executed when a file is selected
//
// Behavior:
//   - Defaults to the user's home directory if baseDir is empty
//   - If allowType is provided, only those file types are selectable
//   - A no‑op callback is used if onSelect is nil
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
