package iplist

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"bgscan/internal/core/fileutil"
)

// IPListDir Directory where IP list files are stored.
const IPListDir = "ips"

// IPFileInfo contains metadata about an IP list file.
type IPFileInfo struct {
	Name      string    // filename without extension
	Path      string    // absolute path
	Size      int64     // file size in bytes
	CreatedAt time.Time // last modification time
}

// getBaseDir resolves the absolute directory where IP lists are stored.
func getBaseDir() (string, error) {
	base, err := fileutil.GetCurrentPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, IPListDir), nil
}

// ListIPFiles returns all CSV files in the IP list directory.
// Results are sorted by modification time (newest first).
func ListIPFiles() ([]IPFileInfo, error) {
	dir, err := getBaseDir()
	if err != nil {
		return nil, err
	}

	files, err := fileutil.ListFiles(
		dir,
		func(name string, _ os.FileInfo) bool {
			return fileutil.HasExt(name, ".csv")
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]IPFileInfo, 0, len(files))

	for _, f := range files {
		out = append(out, IPFileInfo{
			Name:      fileutil.StripExt(f.Name),
			Path:      f.Path,
			Size:      f.Info.Size(),
			CreatedAt: f.Info.ModTime(),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out, nil
}

// GetIPFileInfo returns metadata for an IP list file.
//
// Accepted inputs:
//   - absolute path
//   - filename ("mylist.csv")
//   - name without extension ("mylist")
func GetIPFileInfo(nameOrPath string) (IPFileInfo, error) {
	var fullPath string

	if filepath.IsAbs(nameOrPath) {
		fullPath = nameOrPath
	} else {
		dir, err := getBaseDir()
		if err != nil {
			return IPFileInfo{}, err
		}

		name := nameOrPath
		if !fileutil.HasExt(name, ".csv") {
			name += ".csv"
		}

		fullPath = filepath.Join(dir, name)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return IPFileInfo{}, fmt.Errorf("ip list file not found: %s", nameOrPath)
		}
		return IPFileInfo{}, fmt.Errorf("stat file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return IPFileInfo{}, fmt.Errorf("not a regular file: %s", fullPath)
	}

	return IPFileInfo{
		Name:      fileutil.StripExt(filepath.Base(fullPath)),
		Path:      fullPath,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// GetIPFilePath resolves the absolute path of an IP list file.
func GetIPFilePath(name string) (string, error) {
	dir, err := getBaseDir()
	if err != nil {
		return "", err
	}

	if !fileutil.HasExt(name, ".csv") {
		name += ".csv"
	}

	return filepath.Join(dir, name), nil
}

// FileExists checks whether an IP list file exists.
func FileExists(name string) bool {
	path, err := GetIPFilePath(name)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}
