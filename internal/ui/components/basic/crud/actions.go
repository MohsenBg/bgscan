package crud

import (
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"

	tea "charm.land/bubbletea/v2"
)

type MsgActionTrigger struct{ ActionType string }

// registerActions builds the name → handler table; only non-nil hooks are
// registered.
func (m *Model[T]) registerActions() {
	if m.hooks.OnSelect != nil {
		m.actions["select"] = actionEntry[T]{needsItem: true, run: m.hooks.OnSelect}
	}
	if m.hooks.OnDelete != nil {
		m.actions["delete"] = actionEntry[T]{needsItem: true, run: m.requestDeletion}
	}
	if m.hooks.OnRename != nil {
		m.actions["rename"] = actionEntry[T]{needsItem: true, run: m.requestRename}
	}
	if m.hooks.OnAdd != nil {
		add := m.hooks.OnAdd
		m.actions["add"] = actionEntry[T]{run: func(T) tea.Cmd { return add() }}
	}
	for _, a := range m.hooks.Custom {
		m.actions[a.Name] = actionEntry[T]{needsItem: a.NeedsItem, run: a.Handler}
	}
}

func (m *Model[T]) configureKeymaps() {
	var keys []table.ActionKey

	if m.hooks.OnSelect != nil {
		keys = append(keys, table.NewKey(
			[]string{env.KeyEnter}, "select", "select item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "select"} },
		))
	}
	if m.hooks.OnAdd != nil {
		keys = append(keys, table.NewKey(
			[]string{"a"}, "add", "create item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "add"} },
		))
	}
	if m.hooks.OnDelete != nil {
		keys = append(keys, table.NewKey(
			[]string{"x"}, "delete", "delete item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "delete"} },
		))
	}
	if m.hooks.OnRename != nil {
		keys = append(keys, table.NewKey(
			[]string{"r"}, "rename", "rename item",
			func() tea.Msg { return MsgActionTrigger{ActionType: "rename"} },
		))
	}
	for _, a := range m.hooks.Custom {
		a := a
		keys = append(keys, table.NewKey(
			a.Key, a.Name, a.Help,
			func() tea.Msg { return MsgActionTrigger{ActionType: a.Name} },
		))
	}

	m.table.SetKeys(keys...)
}
