package ipviewer

import (
	"fmt"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		if cmd := m.copySelectedIP(); cmd != nil {
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) copySelectedIP() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return nil
	}

	row, ok := t.SelectedRow()
	if !ok {
		return nil
	}

	col := m.ipColumn()
	if col >= len(row) {
		return nil
	}

	if err := clipboard.WriteAll(row[col]); err != nil {
		return m.errorCmd("Error Copying IP", err.Error())
	}
	return m.infoCmd("IP Copied", fmt.Sprintf("IP copied to clipboard:%s", row[col]))
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.env, title, message, notice.NOTICE_ERROR)
}

func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.env, title, message, notice.NOTICE_INFO)
}
