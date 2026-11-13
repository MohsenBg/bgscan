package fileutil

import (
	"os"
	"path/filepath"
)

// FileEntry is one file found while listing a directory.
type FileEntry struct {
	Name string      // base name, no directory part
	Path string      // absolute path
	Info os.FileInfo // size and modification time
}

// ListFiles returns the regular files in dir.
//
// Directories are skipped, and so are entries whose metadata cannot be read.
// filter, when set, decides which names to keep. Every Path comes back
// absolute, which is what callers store and pass around later.
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
