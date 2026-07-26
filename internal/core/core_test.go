package core

import "testing"

func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
}
