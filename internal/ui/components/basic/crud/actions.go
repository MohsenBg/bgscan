package crud

import (
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"

	tea "charm.land/bubbletea/v2"
)

// MsgActionTrigger asks the model to run the verb with this name. It is the
// public form of "press the key bound to this verb".
type MsgActionTrigger struct{ ActionType string }

// actionEntry is one CRUD verb: the binding that triggers it, whether it
// needs the selected row, and what it runs.
type actionEntry[T any] struct {
	name      string
	binding   keys.Binding
	needsItem bool
	run       func(item T) tea.Cmd
}

// bindActions builds the verb table from the hooks.
//
// Only hooks that are actually set produce an entry, so a read-only source
// advertises no verbs and reacts to no verb keys. Custom verbs keep the
// key strings their provider asked for, because those are not part of the
// shared bindings.
func (m *Model[T]) bindActions(k keys.Keys) {
	m.actions = m.actions[:0]

	if m.hooks.OnSelect != nil {
		m.actions = append(m.actions, actionEntry[T]{
			name:      "select",
			binding:   k.Crud.Select,
			needsItem: true,
			run:       m.hooks.OnSelect,
		})
	}

	if m.hooks.OnAdd != nil {
		add := m.hooks.OnAdd

		m.actions = append(m.actions, actionEntry[T]{
			name:    "add",
			binding: k.Crud.Add,
			run:     func(T) tea.Cmd { return add() },
		})
	}

	if m.hooks.OnDelete != nil {
		m.actions = append(m.actions, actionEntry[T]{
			name:      "delete",
			binding:   k.Crud.Delete,
			needsItem: true,
			run:       m.requestDeletion,
		})
	}

	if m.hooks.OnRename != nil {
		m.actions = append(m.actions, actionEntry[T]{
			name:      "rename",
			binding:   k.Crud.Rename,
			needsItem: true,
			run:       m.requestRename,
		})
	}

	for _, a := range m.hooks.Custom {
		m.actions = append(m.actions, actionEntry[T]{
			name: a.Name,
			binding: keys.Binding{
				Keys:      a.Key,
				HelpText:  a.Help,
				ShortText: a.Name,
			},
			needsItem: a.NeedsItem,
			run:       a.Handler,
		})
	}
}

// shortHints lists the verbs this source supports.
func (m *Model[T]) shortHints() keys.Hints {
	hints := make(keys.Hints, 0, len(m.actions))

	for _, a := range m.actions {
		hints = append(hints, a.binding.ShortItem())
	}

	return hints
}

// longHints is shortHints with every bound key spelled out.
func (m *Model[T]) longHints() keys.Hints {
	hints := make(keys.Hints, 0, len(m.actions))

	for _, a := range m.actions {
		hints = append(hints, a.binding.LongItem())
	}

	return hints
}

// handleAction runs the verb bound to key, resolving the selected row
// first when the verb needs one.
func (m *Model[T]) handleAction(key tea.KeyMsg) (tea.Cmd, bool) {
	for _, a := range m.actions {
		if !a.binding.Matches(key) {
			continue
		}

		var item T

		if a.needsItem {
			selected, err := m.getSelected()
			if err != nil {
				return m.notice("Selection", err.Error()), true
			}

			item = selected
		}

		m.env.Log.Dump("item", item)
		return a.run(item), true
	}

	return nil, false
}

// dispatchAction runs the verb called name.
func (m *Model[T]) dispatchAction(name string) tea.Cmd {
	for _, a := range m.actions {
		if a.name != name {
			continue
		}

		var item T

		if a.needsItem {
			selected, err := m.getSelected()
			if err != nil {
				return m.notice("Selection", err.Error())
			}

			item = selected
		}

		return a.run(item)
	}

	return nil
}
