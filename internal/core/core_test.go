package core

import "testing"

// TestInit checks that Init registers all built-in probe schemas
// without returning an error.
func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
}
