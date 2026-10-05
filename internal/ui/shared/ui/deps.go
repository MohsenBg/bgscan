package ui

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Deps bundles what leaf components need to render and log.
// Layout is shared and read-only: do not mutate it, read sizes via
// BodyContentWidth/Height. Theme is a snapshot taken at construction and
// must be refreshed via SetTheme on ThemeChangedMsg. Pointer sharing alone
// is not enough because huh/bubbles cache lipgloss styles at build time.
type Deps struct {
	Layout *layout.Layout
	Theme  *theme.Theme
	Log    logger.Set
	Keys   *keys.Keys
}

// Deps snapshots shared state for leaf construction.
func (s *AppState) Deps() Deps {
	return Deps{
		Layout: s.Layout,
		Theme:  s.Theme(),
		Log:    s.Log,
		Keys:   keys.Default(),
	}
}
