package workspace

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Layout.Update(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		if len(m.dialog) > 0 {
			lastIdx := len(m.dialog) - 1
			top := m.dialog[lastIdx]

			if env.IsBackKey(msg, top.Mode()) || env.IsQuitKey(msg, top.Mode()) {
				cmds = append(cmds, top.OnClose())
				delete(m.dialogPlacements, top.ID())
				m.dialog[lastIdx] = nil
				m.dialog = m.dialog[:lastIdx]
				return m, tea.Batch(cmds...)
			}
		}

	case dialog.OpenDialogMsg:
		m.dialog = append(m.dialog, msg.Component)
		m.dialogPlacements[msg.Component.ID()] = &dialogOptions{
			XPos:    msg.XPos,
			YPos:    msg.YPos,
			XOffset: msg.XOffset,
			YOffset: msg.YOffset,
			OnClose: msg.OnClose,
		}
		return m, msg.Component.Init()

	case ui.CloseComponentMsg:
		for i, ov := range m.dialog {
			if ov.ID() != msg.ID {
				continue
			}

			if opts := m.dialogPlacements[msg.ID]; opts != nil && opts.OnClose != nil {
				cmds = append(cmds, opts.OnClose)
			}

			m.dialog = append(m.dialog[:i], m.dialog[i+1:]...)
			delete(m.dialogPlacements, msg.ID)
			break
		}
	}

	if len(m.dialog) > 0 {
		lastIdx := len(m.dialog) - 1
		newLayer, cmd := m.dialog[lastIdx].Update(msg)
		m.dialog[lastIdx] = newLayer

		if _, ok := msg.(tea.KeyMsg); ok {
			return m, cmd
		}
		cmds = append(cmds, cmd)
	}

	var hCmd, bCmd, fCmd tea.Cmd
	m.header, hCmd = m.header.Update(msg)
	m.body, bCmd = m.body.Update(msg)
	m.footer, fCmd = m.footer.Update(msg)
	cmds = append(cmds, hCmd, bCmd, fCmd)

	return m, tea.Batch(cmds...)
}
