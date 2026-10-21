package crud

import (
	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
)

// TriggerActionMsg requests the model to execute a specific action by name.
type TriggerActionMsg struct{ Action string }

type actionEntry[T any] struct {
	name      string
	binding   keys.Binding
	needsItem bool
	run       func(item T) tea.Cmd
}

func (m *Model[T]) bindActions(k keys.Keys) {
	m.actions = m.actions[:0]

	if hook := m.hooks.OnSelect; hook != nil {
		m.actions = append(m.actions, actionEntry[T]{
			name:      "select",
			binding:   k.Crud.Select,
			needsItem: true,
			run:       hook,
		})
	}
	if hook := m.hooks.OnAdd; hook != nil {
		m.actions = append(m.actions, actionEntry[T]{
			name:    "add",
			binding: k.Crud.Add,
			run:     func(T) tea.Cmd { return hook() },
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

func (m *Model[T]) shortHints() keys.Hints {
	hints := make(keys.Hints, 0, len(m.actions))
	for _, a := range m.actions {
		hints = append(hints, a.binding.ShortItem())
	}
	return hints
}

func (m *Model[T]) longHints() keys.Hints {
	hints := make(keys.Hints, 0, len(m.actions))
	for _, a := range m.actions {
		hints = append(hints, a.binding.LongItem(m.source.Name()))
	}
	return hints
}

// resolveItem fetches the currently selected item if the action requires one.
func (m *Model[T]) resolveItem(needsItem bool) (T, error) {
	var zero T
	if !needsItem {
		return zero, nil
	}
	return m.getSelected()
}

func (m *Model[T]) handleAction(key tea.KeyMsg) (tea.Cmd, bool) {
	for _, a := range m.actions {
		if !a.binding.Matches(key) {
			continue
		}

		item, err := m.resolveItem(a.needsItem)
		if err != nil {
			return m.showNotice("Selection", err.Error(), notice.NOTICE_INFO), true
		}
		return a.run(item), true
	}
	return nil, false
}

func (m *Model[T]) dispatchAction(name string) tea.Cmd {
	for _, a := range m.actions {
		if a.name != name {
			continue
		}

		item, err := m.resolveItem(a.needsItem)
		if err != nil {
			return m.showNotice("Selection", err.Error(), notice.NOTICE_INFO)
		}
		return a.run(item)
	}
	return nil
}
