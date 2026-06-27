package app

import (
	"bgscan/internal/ui/main/body"
	"bgscan/internal/ui/main/footer"
	"bgscan/internal/ui/main/header"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// OverlayPlacement defines how an overlay component
// should be positioned relative to the terminal layout.
type OverlayPlacement struct {
	XPos    ui.OverlayPosition
	YPos    ui.OverlayPosition
	XOffset int
	YOffset int
}

// model is the root BubbleTea application model.
// It coordinates layout updates, base UI components,
// and overlay layers.
type model struct {
	layout *layout.Layout

	// layers contains active overlay components.
	layers []ui.Component

	// overlayPlacements stores the position metadata
	// for each overlay component.
	overlayPlacements map[ui.ComponentID]*OverlayPlacement

	header ui.Component
	body   ui.Component
	footer ui.Component
}

// New creates and initializes the root application model.
func New() tea.Model {
	l := layout.New()

	return &model{
		layout:            l,
		layers:            make([]ui.Component, 0, 5),
		overlayPlacements: make(map[ui.ComponentID]*OverlayPlacement),
		header:            header.New(l),
		body:              body.New(l),
		footer:            footer.New(l),
	}
}

// Init initializes all base UI components.
func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.header.Init(),
		m.body.Init(),
		m.footer.Init(),
	)
}

// getOverlayPlacement returns the placement configuration
// for an overlay component. If no placement exists,
// a default centered placement is created and stored.
func (m *model) getOverlayPlacement(id ui.ComponentID) *OverlayPlacement {
	if p, ok := m.overlayPlacements[id]; ok {
		return p
	}

	p := &OverlayPlacement{
		XPos:    ui.Center,
		YPos:    ui.Center,
		XOffset: 0,
		YOffset: 0,
	}

	m.overlayPlacements[id] = p
	return p
}
