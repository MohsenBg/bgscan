package table

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		// In filter mode all keystrokes edit the filter (never nav/sort/
		// help), so typing can't move the cursor. Esc lands here too when
		// idle: a harmless clear that refreshes the unfiltered view.
		if m.filtering || msg.Key().Code == tea.KeyEsc {
			m.updateFilterLocked(msg)
			return m, nil
		}

		if cmd := m.Keys.Check(msg); cmd != nil {
			cmds = append(cmds, cmd)
		} else {
			// Sort/filter keys run only when no action consumed the
			// key, so a provider-registered action would win.
			switch msg.String() {
			case "s":
				m.cycleSortColumnLocked()
			case "o":
				m.toggleSortDirectionLocked()
			case "/":
				m.filtering = true
			}
		}

		if msg.String() == "?" {
			m.FullHelp = !m.FullHelp
			m.updateTableSizeLocked()
		}

	case tea.WindowSizeMsg:
		m.updateTableSizeLocked()
		m.BubbleTable.SetStyles(m.tableStyles())
		return m, nil
	}

	var tableCmd tea.Cmd
	m.BubbleTable, tableCmd = m.BubbleTable.Update(msg)
	if tableCmd != nil {
		cmds = append(cmds, tableCmd)
	}

	return m, tea.Batch(cmds...)
}

// updateFilterLocked edits the filter from keystrokes: printable text
// appends and re-filters live (cursor kept), backspace deletes, enter keeps
// the filter and exits edit mode, esc clears it and exits. Caller holds m.mu.
func (m *Model) updateFilterLocked(msg tea.Msg) {
	kp, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return
	}
	switch kp.Code {
	case tea.KeyEsc:
		m.filter, m.filtering = "", false
		m.refreshDisplayLocked(false)
	case tea.KeyEnter:
		m.filtering = false
	case tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
			m.refreshDisplayLocked(false)
		}
	default:
		if kp.Text != "" {
			m.filter += kp.Text
			m.refreshDisplayLocked(false)
		}
	}
}
