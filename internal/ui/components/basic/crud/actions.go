package crud

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"

	tea "charm.land/bubbletea/v2"
)

type MsgActionTrigger struct{ ActionType string }

func (m *Model[T]) configureKeymaps() {
	var keys []table.ActionKey

	if _, ok := m.provider.OnSelect(m.zeroValue()); ok {
		keys = append(keys, table.NewKey(
			[]string{env.KeyEnter},
			"select",
			"select item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "select"} },
		))
	}

	if m.CanAdd() {
		keys = append(keys, table.NewKey(
			[]string{"a"},
			"add",
			"create item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "add"} },
		))
	}

	if _, ok := m.provider.OnDelete(m.zeroValue()); ok {
		keys = append(keys, table.NewKey(
			[]string{"x"},
			"delete",
			"delete item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "delete"} },
		))
	}

	if _, ok := m.provider.OnRename(m.zeroValue(), ""); ok {
		keys = append(keys, table.NewKey(
			[]string{"r"},
			"rename",
			"rename item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "rename"} },
		))
	}

	m.table.SetKeys(keys...)
}

func (m *Model[T]) zeroValue() T {
	var zero T
	return zero
}
