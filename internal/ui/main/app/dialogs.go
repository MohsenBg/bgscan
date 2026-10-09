package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/anim"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

type frameOptions = dialog.OpenDialogMsg

// layer is an open dialog and its transition state.
// Closing layers stay in the stack until their collapse finishes.
type layer struct {
	component ui.Component
	opts      *frameOptions

	openedAt  time.Time
	closingAt time.Time
	closeCmd  tea.Cmd
}

func (l *layer) closing() bool {
	return !l.closingAt.IsZero()
}

// progress returns 0 at the start of the current transition and 1 at the end.
// Closing reuses the open curve in reverse.
func (l *layer) progress(now time.Time) float64 {
	if l.closing() {
		return 1 - fraction(l.closingAt, now, dialog.Duration)
	}

	return fraction(l.openedAt, now, dialog.Duration)
}

// settled reports whether the layer has no transition left to play.
func (l *layer) settled(now time.Time) bool {
	at := l.openedAt
	if l.closing() {
		at = l.closingAt
	}

	return at.IsZero() || now.Sub(at) >= dialog.Duration
}

// frame returns the layer's frame options, initializing them on first use.
func (l *layer) frame() *frameOptions {
	if l.opts == nil {
		l.opts = &frameOptions{
			XPos: dialog.Center,
			YPos: dialog.Center,
		}
	}

	return l.opts
}

// fraction returns how much of d has passed since at, clamped to [0, 1].
func fraction(at, now time.Time, d time.Duration) float64 {
	if at.IsZero() || d <= 0 {
		return 1
	}

	return anim.Clamp01(now.Sub(at).Seconds() / d.Seconds())
}

// dialogFrameMsg paces dialog transitions.
type dialogFrameMsg time.Time

// tickDialogFrames schedules the next transition frame.
func tickDialogFrames() tea.Cmd {
	return tea.Tick(dialog.FrameInterval, func(t time.Time) tea.Msg {
		return dialogFrameMsg(t)
	})
}

type dialogStack struct {
	layers []*layer
}

func newDialogStack() dialogStack {
	return dialogStack{
		layers: make([]*layer, 0, 5),
	}
}

// top returns the active dialog layer, or nil when none is open.
func (s *dialogStack) top() *layer {
	if n := len(s.layers); n > 0 {
		return s.layers[n-1]
	}

	return nil
}

// push adds a dialog to the top of the stack and starts its open transition.
func (s *dialogStack) push(msg dialog.OpenDialogMsg) {
	setInDialog(msg.Component, true)

	s.layers = append(s.layers, &layer{
		component: msg.Component,
		opts:      &msg,
		openedAt:  time.Now(),
	})
}

// beginCloseTop starts the close transition of the active dialog.
func (s *dialogStack) beginCloseTop() tea.Cmd {
	l := s.top()
	if l == nil || l.closing() {
		return nil
	}

	return s.beginClose(l, l.component.OnClose())
}

// beginCloseByID starts the close transition of one dialog.
func (s *dialogStack) beginCloseByID(id ui.ComponentID) tea.Cmd {
	for _, l := range s.layers {
		if l.component.ID() != id {
			continue
		}

		return s.beginClose(l, l.frame().OnClose)
	}

	return nil
}

// beginClose marks a layer as leaving and stores the command to run after
// the close transition finishes.
func (s *dialogStack) beginClose(l *layer, cmd tea.Cmd) tea.Cmd {
	if l.closing() {
		return nil
	}

	l.closingAt = time.Now()
	l.closeCmd = cmd

	return tickDialogFrames()
}

// expire removes layers whose close transition finished and returns their
// deferred OnClose commands.
func (s *dialogStack) expire(now time.Time) []tea.Cmd {
	var cmds []tea.Cmd

	kept := s.layers[:0]

	for _, l := range s.layers {
		if l.closing() && l.settled(now) {
			setInDialog(l.component, false)

			if l.closeCmd != nil {
				cmds = append(cmds, l.closeCmd)
			}

			continue
		}

		kept = append(kept, l)
	}

	// Release the tail so removed layers are not pinned in memory.
	for i := len(kept); i < len(s.layers); i++ {
		s.layers[i] = nil
	}

	s.layers = kept

	return cmds
}

// animating reports whether any layer is still transitioning.
func (s *dialogStack) animating(now time.Time) bool {
	for _, l := range s.layers {
		if !l.settled(now) {
			return true
		}
	}

	return false
}

// updateTop sends msg to the active dialog.
// Closing dialogs swallow input so keys do not reach the screen behind them.
func (s *dialogStack) updateTop(msg tea.Msg) (tea.Cmd, bool) {
	l := s.top()
	if l == nil {
		return nil, false
	}

	if l.closing() {
		return nil, true
	}

	var cmd tea.Cmd
	l.component, cmd = l.component.Update(msg)

	return cmd, true
}

// updateAll sends msg to every layer, including closing ones, so resize and
// theme updates remain consistent during transitions.
func (s *dialogStack) updateAll(msg tea.Msg) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(s.layers))

	for _, l := range s.layers {
		var cmd tea.Cmd
		l.component, cmd = l.component.Update(msg)

		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}

// setInDialog tells a component whether it is hosted inside a dialog.
func setInDialog(c ui.Component, inDialog bool) {
	if f, ok := c.(interface{ SetInDialog(bool) }); ok {
		f.SetInDialog(inDialog)
	}
}
