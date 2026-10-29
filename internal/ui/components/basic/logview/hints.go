package logview

import (
	"charm.land/bubbles/v2/viewport"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
)

// helpSet builds the short and long help for the log viewer from the shared
// key bindings, so the bar always matches what the keys actually do.
func helpSet(k *keys.Keys) help.HelpSet {
	lv := k.LogView
	g := k.Global

	short := keys.Hints{
		keys.Combine("move", lv.Up, lv.Down),
		keys.Combine("page", lv.PageUp, lv.PageDown),
	}
	short = append(short, keys.Short(g.Back, g.Quit, g.Help)...)

	long := keys.Long(
		lv.Up,
		lv.Down,
		lv.PageUp,
		lv.PageDown,
		lv.HalfPageUp,
		lv.HalfPageDown,
		lv.Top,
		lv.Bottom,
	)
	long = append(long, keys.Long(g.Back, g.Quit, g.Help)...)

	return help.HelpSet{Short: short, Long: long}
}

// viewportKeyMap points the viewport at the shared log view bindings instead
// of its own defaults (which would also grab "b", "f" and "d").
func viewportKeyMap(k keys.LogViewKeys) viewport.KeyMap {
	km := viewport.DefaultKeyMap()
	km.Up.SetKeys(k.Up.Keys...)
	km.Down.SetKeys(k.Down.Keys...)
	km.PageUp.SetKeys(k.PageUp.Keys...)
	km.PageDown.SetKeys(k.PageDown.Keys...)
	km.HalfPageUp.SetKeys(k.HalfPageUp.Keys...)
	km.HalfPageDown.SetKeys(k.HalfPageDown.Keys...)
	return km
}
