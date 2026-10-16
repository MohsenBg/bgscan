package inspector

import "github.com/MohsenBg/bgscan/internal/ui/shared/keys"

func (m *Model) ShortHints() keys.Hints {
	k := m.env.Keys.Inspector
	g := m.env.Keys.Global

	hints := keys.Hints{
		keys.Combine("move", k.Up, k.Down),
		keys.Combine("switch tab", k.NextTab, k.PrevTab),
	}
	hints = append(hints, keys.Short(k.Edit, k.Description)...)
	return append(hints, keys.Short(g.Back, g.Quit, g.Help)...)
}

func (m *Model) LongHints() keys.Hints {
	k := m.env.Keys.Inspector
	g := m.env.Keys.Global
	return keys.Long(
		k.Up, k.Down, k.NextTab, k.PrevTab,
		k.Edit, k.Description,
		g.Back, g.Quit, g.Help,
	)
}

func DialogHints(k keys.InspectorKeys) keys.Hints {
	hints := keys.Hints{
		keys.Combine("move", k.Up, k.Down),
	}
	return append(hints, keys.Short(k.Edit, k.Description)...)
}
