package result

import (
	"log"
	"testing"

	"bgscan/internal/logger"
)

func TestMain(m *testing.M) {
	if err := logger.InitCore(); err != nil {
		log.Fatalf("core logger initialization failed: %v", err)
	}
	m.Run()
}
