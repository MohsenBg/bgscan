package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// FormatMarker is the file that records a directory's layout version.
const FormatMarker = ".format"

// ReadVersion returns the layout version recorded in dir.
// ReadVersion reports found as false with no error when dir has no
// marker file. It returns an error when dir is empty, the marker is
// empty or malformed, or the stored version is below 1.
func ReadVersion(dir string) (version int, found bool, err error) {
	if dir == "" {
		return 0, false, fmt.Errorf("read format marker: directory is required")
	}

	path := filepath.Join(dir, FormatMarker)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("read format marker %q: %w", path, err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return 0, true, fmt.Errorf("format marker %q is empty", path)
	}

	version, err = strconv.Atoi(text)
	if err != nil {
		return 0, true, fmt.Errorf("parse format marker %q: %w", path, err)
	}
	if version < 1 {
		return 0, true, fmt.Errorf("format marker %q holds invalid version %d", path, version)
	}

	return version, true, nil
}

// WriteVersion records version in dir, creating the directory if needed.
// WriteVersion returns an error when dir is empty or version is below 1.
// The marker file is written with an atomic replace.
func WriteVersion(dir string, version int) error {
	if dir == "" {
		return fmt.Errorf("write format marker: directory is required")
	}
	if version < 1 {
		return fmt.Errorf("write format marker: version must be positive, got %d", version)
	}

	if err := fileutil.EnsureDir(dir); err != nil {
		return err
	}

	path := filepath.Join(dir, FormatMarker)
	data := []byte(strconv.Itoa(version) + "\n")

	if err := fileutil.WriteFileAtomic(path, data, fileutil.DefaultFileMode); err != nil {
		return fmt.Errorf("write format marker %q: %w", path, err)
	}

	return nil
}
