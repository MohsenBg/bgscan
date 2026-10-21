package body

import (
	"slices"
	"time"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/main/footer"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	n := len(m.screens)
	if n == 0 {
		return m, nil
	}

	top := m.screens[n-1].component
	env := m.session.Env()

	switch msg := msg.(type) {
	case theme.ThemeChanged:
		cmds := make([]tea.Cmd, 0, n)
		for i := range m.screens {
			component, cmd := m.screens[i].component.Update(msg)
			m.screens[i].component = component
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case bodyFrameMsg:
		now := time.Now()
		removals, dropped := m.expire(now)

		cmds := removals
		if m.swap != nil {
			if m.swap.settled(now) {
				m.swap = nil
			} else {
				cmds = append(cmds, tickSwaps())
			}
		}

		if dropped {
			cmds = append(cmds, m.stackChangedCmd())
		}

		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		if m.swap != nil && !m.swap.settled(time.Now()) {
			return m, nil
		}

		keys := env.Keys
		switch top.Mode() {
		case ui.KeyModeNormal:
			if keys.Global.Back.Matches(msg) {
				if n > 1 {
					return m.popComponent()
				}
				return m, confirm.ExitConfirmCmd(env)
			}

			if keys.Global.Quit.Matches(msg) {
				return m, confirm.ExitConfirmCmd(env)
			}

		case ui.KeyModeInput:
			if keys.Global.BackFromInput.Matches(msg) {
				return m.popComponent()
			}
		}

	case ui.OpenComponentMsg:
		if msg.Component == nil {
			break
		}

		env.Log.Info("Opening screen: %s", msg.Component.Name())
		return m.pushComponent(msg.Component)

	case ui.CloseComponentMsg:
		for i := range m.screens {
			s := m.screens[i]
			if s.component.ID() != msg.ID {
				continue
			}
			if i == 0 || s.closing() {
				return m, nil
			}
			if i == n-1 && env.Settings.Animated {
				return m.popComponent()
			}

			cmd := s.component.OnClose()
			m.screens = slices.Delete(m.screens, i, i+1)
			return m, tea.Batch(cmd, m.stackChangedCmd())
		}

	case ui.ResetComponentStacksMsg:
		m.swap = nil
		cmds := make([]tea.Cmd, 0, n)

		for i := 1; i < n; i++ {
			cmds = append(cmds, m.screens[i].component.OnClose())
			m.screens[i] = screen{}
		}

		m.screens = m.screens[:1]
		cmds = append(cmds, m.stackChangedCmd())
		return m, tea.Batch(cmds...)
	}

	component, cmd := top.Update(msg)
	m.screens[n-1].component = component

	return m, cmd
}

// expire removes screens whose closing transitions have finished.
func (m *Model) expire(now time.Time) ([]tea.Cmd, bool) {
	kept := m.screens[:0]
	var (
		cmds    []tea.Cmd
		dropped bool
	)

	for i, s := range m.screens {
		if s.closing() && s.settled(now) {
			cmds = append(cmds, s.component.OnClose())
			m.screens[i] = screen{}
			dropped = true
			continue
		}

		kept = append(kept, s)
	}

	clear(m.screens[len(kept):])
	m.screens = kept

	return cmds, dropped
}

func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	cmds := m.settleSwap()

	var leaving ui.Component
	if n := len(m.screens); n > 0 {
		leaving = m.screens[n-1].component
	}

	m.screens = append(m.screens, screen{component: c})
	cmds = append(cmds, c.Init(), m.forceResize(), m.stackChangedCmd())

	if leaving != nil && m.session.Env().Settings.Animated {
		m.swap = &swap{
			leaving:  leaving,
			entering: c,
			kind:     pushOpen,
			at:       time.Now(),
		}
		cmds = append(cmds, tickSwaps())
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	n := len(m.screens)
	if n <= 1 {
		return m, nil
	}

	top := m.screens[n-1]
	if top.closing() {
		return m, nil
	}

	if !m.session.Env().Settings.Animated {
		cmd := top.component.OnClose()
		m.screens[n-1] = screen{}
		m.screens = m.screens[:n-1]

		return m, tea.Batch(cmd, m.stackChangedCmd())
	}

	cmds := m.settleSwap()
	now := time.Now()

	m.screens[n-1] = screen{
		component: top.component,
		closingAt: now,
	}
	m.swap = &swap{
		leaving:  top.component,
		entering: m.screens[n-2].component,
		kind:     pushClose,
		at:       now,
	}

	cmds = append(cmds, tickSwaps())
	return m, tea.Batch(cmds...)
}

// settleSwap completes the current transition and processes expired screens.
func (m *Model) settleSwap() []tea.Cmd {
	if m.swap == nil {
		return nil
	}

	removals, dropped := m.expire(time.Now())
	cmds := removals
	m.swap = nil

	if dropped {
		cmds = append(cmds, m.stackChangedCmd())
	}

	return cmds
}

func (m *Model) topName() string {
	if len(m.screens) == 0 {
		return ""
	}

	return m.screens[len(m.screens)-1].component.Name()
}

func (m *Model) stackChangedCmd() tea.Cmd {
	name := m.topName()
	return func() tea.Msg {
		return footer.UpdateStatus{Status: name}
	}
}

func (m *Model) forceResize() tea.Cmd {
	term := m.session.Env().Layout.Terminal()
	msg := tea.WindowSizeMsg{
		Width:  term.Width,
		Height: term.Height,
	}

	return func() tea.Msg {
		return msg
	}
}
