package table

import "github.com/MohsenBg/bgscan/internal/ui/shared/keys"

func (m *Model) longHits(extra ...keys.Hint) keys.Hints {
	k := m.env.Keys.Table
	g := m.env.Keys.Global

	hints := keys.Long(
		k.Up,
		k.Down,
		k.PageUp,
		k.PageDown,
		k.HalfPageUp,
		k.HalfPageDown,
		k.Top,
		k.Bottom,
		k.ChangeSort,
		k.ToggleOrder,
		k.Filter,
	)

	hints = append(hints, extra...)

	return append(hints, keys.Long(g.Back, g.Quit, g.Help)...)
}

func (m *Model) shortHits(extra ...keys.Hint) keys.Hints {
	k := m.env.Keys.Table
	g := m.env.Keys.Global

	hints := keys.Hints{
		keys.Combine("move", k.Up, k.Down),
	}

	hints = append(hints, keys.Short(
		k.ChangeSort,
		k.ToggleOrder,
		k.Filter,
	)...)

	hints = append(hints, extra...)

	return append(hints, keys.Short(g.Back, g.Quit, g.Help)...)
}
