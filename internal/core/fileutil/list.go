package fileutil

import (
	"os"
	"path/filepath"
)

// ═══════════════════════════════════════════════════════════
// List Files Operations
// ═══════════════════════════════════════════════════════════

// FileEntry represents a discovered file in a directory listing.
type FileEntry struct {
	Name string      // file name only
	Path string      // absolute file path
	Info os.FileInfo // file metadata
}

// ListFiles lists files inside a directory (non-recursive).
// Directories are skipped. An optional filter can be provided to
// control which files are returned.
func ListFiles(
	dir string,
	filter func(name string, info os.FileInfo) bool,
) ([]FileEntry, error) {

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]FileEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		info, err := entry.Info()
		if err != nil {
			continue // skip unreadable entries
		}

		if filter != nil && !filter(name, info) {
			continue
		}

		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			continue
		}

		files = append(files, FileEntry{
			Name: name,
			Path: absPath,
			Info: info,
		})
	}

	return files, nil
}
