package result

import (
	"log"
	"os"
	"testing"
)

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
