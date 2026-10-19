package theme

import (
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/meter"
)

func TestTierColorFallback(t *testing.T) {
	// Zero-value theme falls back to generic tokens.
	var empty Theme
	if got := empty.TierColor(meter.TierExcellent); got != empty.Success {
		t.Errorf("fallback: got %v, want Success", got)
	}
	if got := empty.TierColor(meter.TierFail); got != empty.Muted {
		t.Errorf("fallback: got %v, want Muted", got)
	}

	// An override wins over the fallback.
	custom := Theme{TierGood: lipgloss.Color("#123456")}
	if got := custom.TierColor(meter.TierGood); got != custom.TierGood {
		t.Errorf("override: got %v, want TierGood", got)
	}
	if got := custom.TierColor(meter.TierFair); got != custom.Warning {
		t.Errorf("override: unmatched tier must fall back, got %v", got)
	}

	// Plain cells get no color.
	if got := empty.TierColor(meter.TierPlain); got != nil {
		t.Errorf("TierPlain must be nil, got %v", got)
	}
}

func TestThemesHaveTierGood(t *testing.T) {
	Init()

	// Every builtin theme seeds TierGood explicitly; the visual spot between
	// Success and Fair is exactly where gray looked broken.
	for _, name := range List() {
		th, err := Get(name)
		if err != nil {
			t.Fatalf("Get(%q): %v", name, err)
		}
		if th.TierGood == nil {
			t.Errorf("theme %q has no TierGood seeded", name)
		}
	}
}
