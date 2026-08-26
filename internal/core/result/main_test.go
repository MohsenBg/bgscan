package result

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"bgscan/internal/logger"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "result-test-*")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	if err := logger.InitCoreToDir(filepath.Join(dir, "logs")); err != nil {
		log.Fatalf("core logger initialization failed: %v", err)
	}
	os.Exit(m.Run())
}
