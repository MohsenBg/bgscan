package fileutil

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
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
