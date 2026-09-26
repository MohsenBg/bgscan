package theme

import (
	"slices"
	"testing"
)

func TestRegistryContainsBuiltins(t *testing.T) {
	Init()

	for _, name := range []string{
		"gruvbox-dark", "gruvbox-light",
		"one-dark", "one-light",
		"dracula",
		"monokai",
		"solarized-dark", "solarized-light",
		"rose-pine", "rose-pine-dawn",
		"kanagawa",
		"github-dark", "github-light",
		"material-dark",
		"everforest-dark", "everforest-light",
		"ayu-dark", "ayu-light",
	} {
		if !slices.Contains(List(), name) {
			t.Errorf("theme %q not registered", name)
		}
	}
}

func TestGetAndUnknown(t *testing.T) {
	Init()

	th, err := Get("dracula")
	if err != nil {
		t.Fatalf("Get(dracula): %v", err)
	}
	if th.Name != "dracula" {
		t.Errorf("Get(dracula).Name = %q, want dracula", th.Name)
	}

	if _, err := Get("bogus-theme"); err == nil {
		t.Fatal("Get(bogus-theme) = nil, want error")
	}

	auto, err := Get("")
	if err != nil {
		t.Fatalf("Get(\u0022\u0022): %v", err)
	}
	if auto.Name != BGScanDark.Name && auto.Name != BGScanLight.Name {
		t.Errorf("Get(\u0022\u0022).Name = %q, want an auto-detected builtin", auto.Name)
	}
}
