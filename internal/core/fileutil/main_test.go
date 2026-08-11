package fileutil

import (
	"log"
	"os"
	"testing"

	"bgscan/internal/logger"
)

func TestMain(m *testing.M) {
	if err := logger.InitCore(); err != nil {
		log.Fatalf("core logger initialization failed: %v", err)
	}

	switch os.Getenv(basePathHelperEnv) {
	case basePathHelperReal, basePathHelperSymlinked:
		base, err := BasePath()
		if err != nil {
			os.Exit(2)
		}
		_, _ = os.Stdout.WriteString(base)
		os.Exit(0)
	}

	os.Exit(m.Run())
}
