package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

type frameOptions = dialog.OpenDialogMsg

type dialogStack struct {
	layers []ui.Component
	frames map[ui.ComponentID]*frameOptions
}

func newDialogStack() dialogStack {
	return dialogStack{
		layers: make([]ui.Component, 0, 5),
		frames: make(map[ui.ComponentID]*frameOptions),
	}
}

// top returns the active dialog, or nil when none is open.
func (s *dialogStack) top() ui.Component {
	if n := len(s.layers); n > 0 {
		return s.layers[n-1]
	}

	return nil
}

// push adds a dialog to the top of the stack.
func (s *dialogStack) push(msg dialog.OpenDialogMsg) {
	setInDialog(msg.Component, true)
	s.layers = append(s.layers, msg.Component)
	s.frames[msg.Component.ID()] = &msg
}

// popTop closes the active dialog and returns its OnClose command.
func (s *dialogStack) popTop() tea.Cmd {
	if len(s.layers) == 0 {
		return nil
	}

	last := len(s.layers) - 1
	top := s.layers[last]

	setInDialog(top, false)
	delete(s.frames, top.ID())

	s.layers[last] = nil
	s.layers = s.layers[:last]

	return top.OnClose()
}

// close removes a dialog by ID and returns its OnClose command.
func (s *dialogStack) close(id ui.ComponentID) tea.Cmd {
	for i, c := range s.layers {
		if c == nil || c.ID() != id {
			continue
		}

		var cmd tea.Cmd
		if opts, ok := s.frames[id]; ok && opts.OnClose != nil {
			cmd = opts.OnClose
		}

		setInDialog(c, false)

		copy(s.layers[i:], s.layers[i+1:])
		s.layers[len(s.layers)-1] = nil
		s.layers = s.layers[:len(s.layers)-1]

		delete(s.frames, id)

		return cmd
	}

	return nil
}

// frameFor returns the frame options for a dialog.
func (s *dialogStack) frameFor(id ui.ComponentID) *frameOptions {
	if opts, ok := s.frames[id]; ok {
		return opts
	}

	opts := &frameOptions{
		XPos: dialog.Center,
		YPos: dialog.Center,
	}

	s.frames[id] = opts

	return opts
}

// updateTop forwards msg to the active dialog.
func (s *dialogStack) updateTop(msg tea.Msg) (tea.Cmd, bool) {
	last := len(s.layers) - 1
	if last < 0 {
		return nil, false
	}

	var cmd tea.Cmd
	s.layers[last], cmd = s.layers[last].Update(msg)

	return cmd, true
}

// updateAll forwards msg to every open dialog.
func (s *dialogStack) updateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(s.layers))

	for i, layer := range s.layers {
		var cmd tea.Cmd
		s.layers[i], cmd = layer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return cmds
}

// setInDialog tells a component whether it is hosted inside a dialog.
func setInDialog(c ui.Component, inDialog bool) {
	if f, ok := c.(interface{ SetInDialog(bool) }); ok {
		f.SetInDialog(inDialog)
	}
}
