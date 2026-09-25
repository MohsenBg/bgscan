package ui

import (
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/theme"
)

// Deps bundles what leaf components need to render and log. Theme is a
// snapshot taken at construction and refreshed via SetTheme on
// ThemeChangedMsg — sharing Deps never replaces that event.
type Deps struct {
	Layout *layout.Layout
	Theme  *theme.Theme
	Log    logger.Set
}

// Deps snapshots shared state for leaf construction.
func (s *AppState) Deps() Deps {
	return Deps{
		Layout: s.Layout,
		Theme:  s.Theme(),
		Log:    s.Log,
	}
}
