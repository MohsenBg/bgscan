package fileutil

import (
	"os"
	"path/filepath"
)

// FileEntry represents a discovered file in a directory listing.
type FileEntry struct {
	Name string
	Path string
	Info os.FileInfo
}

// ListFiles returns the non-directory entries in dir.
// When non-nil, filter determines which entries are included.
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
