package formkit

import (
	"fmt"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/form"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/inspector"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Option configures a Base at construction time.
type Option func(*Base)

// WithMaxHeight overrides the default maximum form body height.
func WithMaxHeight(h int) Option {
	return func(b *Base) {
		if h > 0 {
			b.maxHeight = h
		}
	}
}

// Base holds the common component state and lifecycle shared by every tunnel
// configuration form. Protocol models embed it and only add their typed
// config plus protocol-specific inspector fields.
type Base struct {
	id           ui.ComponentID
	layout       *layout.Layout
	state        *ui.AppState
	form         *form.Model
	inspector    *inspector.Model
	name         string
	originalName string
	width        int
	height       int
	maxWidth     int
	maxHeight    int
}

// NewBase creates the shared form base for the given config identity.
func NewBase(
	l *layout.Layout,
	state *ui.AppState,
	name, originalName string,
	opts ...Option,
) Base {
	b := Base{
		id:           ui.NewComponentID(),
		layout:       l,
		state:        state,
		name:         name,
		originalName: originalName,
		maxWidth:     MaxFormWidth,
		maxHeight:    MaxFormHeight,
	}

	for _, opt := range opts {
		opt(&b)
	}

	b.CalculateSize()
	return b
}

// Layout returns the layout used for sizing and positioning.
func (b *Base) Layout() *layout.Layout { return b.layout }

// State returns the application state the form was created with.
func (b *Base) State() *ui.AppState { return b.state }

// Name returns the config name being created or edited.
func (b *Base) Name() string { return b.name }

// SetName sets the config name.
func (b *Base) SetName(name string) { b.name = name }

// ID returns the unique component identifier.
func (b *Base) ID() ui.ComponentID { return b.id }

// Mode returns the UI input mode for managed config forms.
func (b *Base) Mode() env.Mode { return env.ManagedMode }

// CalculateSize derives the form body size from the layout.
func (b *Base) CalculateSize() {
	w := b.layout.BodyContentWidth() - FormPadding
	h := b.layout.BodyContentHeight() - FormPadding

	b.width = min(w, b.maxWidth)
	b.height = min(h, b.maxHeight)
}

// Form returns the underlying form component, or nil before BuildForm.
func (b *Base) Form() *form.Model { return b.form }

// Inspector returns the underlying field inspector, or nil before BuildForm.
func (b *Base) Inspector() *inspector.Model { return b.inspector }

// BuildForm assembles the form component around ins, wiring the shared save
// confirmation, cancel handler and the config-derived validation.
func (b *Base) BuildForm(
	title string,
	ins *inspector.Model,
	validate func() map[string]error,
	onSave func() tea.Msg,
) {
	b.inspector = ins

	b.form = form.New(
		b.layout,
		ins,
		form.WithName(title),
		form.WithWidth(b.width),
		form.WithHeight(b.height),
		form.WithValidation(func(fm *form.Model) error {
			errs := validate()
			if len(errs) == 0 {
				return nil
			}
			return fmt.Errorf("%s", form.FormatValidationErrors(errs))
		}),
		form.WithSave(confirm.ConfirmCmd(
			b.layout,
			"Save configuration?",
			onSave,
			true,
		)),
		form.WithCancel(b.CancelCmd()),
	)
}

// Refresh returns a command that redraws the inspector (used after fields
// change the visibility of other fields).
func (b *Base) Refresh() tea.Cmd {
	if b.inspector != nil {
		return b.inspector.Refresh()
	}
	return nil
}

// Init implements [ui.Component].
func (b *Base) Init() tea.Cmd {
	if b.form == nil {
		return nil
	}
	return b.form.Init()
}

// OnClose implements [ui.Component].
func (b *Base) OnClose() tea.Cmd {
	if b.form == nil {
		return nil
	}
	return b.form.OnClose()
}

// View renders the form.
func (b *Base) View() string {
	if b.form == nil {
		return ""
	}
	return b.form.View()
}

// Update implements [ui.Component]. self is returned so promoted protocol
// models keep their concrete type in the component stack.
func (b *Base) Update(self ui.Component, msg tea.Msg) (ui.Component, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		b.CalculateSize()
		if b.form != nil {
			b.form.SetWidth(b.width)
			b.form.SetHeight(b.height)
		}
	}

	if b.form == nil {
		return self, nil
	}

	component, cmd := b.form.Update(msg)
	b.form = component.(*form.Model)
	return self, cmd
}

// CancelCmd returns the discard-changes confirmation used by the form cancel
// handler.
func (b *Base) CancelCmd() tea.Cmd {
	return confirm.ConfirmCmd(
		b.layout,
		"Discard unsaved changes?",
		func() tea.Msg {
			return ui.CloseComponentMsg{ID: b.ID()}
		},
		false,
	)
}
