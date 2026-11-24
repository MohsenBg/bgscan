package result

import (
	"log"
	"os"
	"testing"
)

// TestMain sets up shared state for the package tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "result-test-*")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	os.Exit(m.Run())
}
