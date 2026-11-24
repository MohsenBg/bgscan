package target

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// List directories below the base path, one per target kind.
var (
	IPListDir     = path.Join("targets", KindIP.Dir())
	DomainListDir = path.Join("targets", KindDomain.Dir())
)

// baseDirOverride points the base directory somewhere else in tests.
// Always empty in production.
var baseDirOverride string

// FileInfo is metadata about one target list file.
type FileInfo struct {
	Kind      Kind
	Name      string
	Path      string
	Size      int64
	CreatedAt time.Time
}

func getBaseRoot() (string, error) {
	if baseDirOverride != "" {
		return baseDirOverride, nil
	}

	return fileutil.BasePath()
}

func getBaseDir(subdir string) (string, error) {
	base, err := getBaseRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, subdir), nil
}

func getIPBaseDir() (string, error) {
	return getBaseDir(IPListDir)
}

func getDomainBaseDir() (string, error) {
	return getBaseDir(DomainListDir)
}

func baseDirForKind(kind Kind) (string, error) {
	switch kind {
	case KindIP:
		return getIPBaseDir()
	case KindDomain:
		return getDomainBaseDir()
	default:
		return "", fmt.Errorf("unsupported target kind: %v", kind)
	}
}

// ListFiles returns all CSV target files, newest first. A missing directory is
// treated as empty so a fresh install does not fail before the first import.
func ListFiles() ([]FileInfo, error) {
	kinds := []Kind{KindIP, KindDomain}

	var out []FileInfo

	for _, kind := range kinds {
		dir, err := baseDirForKind(kind)
		if err != nil {
			return nil, err
		}

		files, err := fileutil.ListFiles(dir, func(name string, _ os.FileInfo) bool {
			return fileutil.HasExt(name, ".csv")
		})
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}

		for _, f := range files {
			out = append(out, FileInfo{
				Kind:      kind,
				Name:      fileutil.StripExt(f.Name),
				Path:      f.Path,
				Size:      f.Info.Size(),
				CreatedAt: f.Info.ModTime(),
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out, nil
}

// validateListName rejects anything that isn't a plain file name, so a list
// name can never escape its target directory.
func validateListName(name string) error {
	if strings.TrimSpace(name) == "" || name == "." || name == ".." ||
		name != filepath.Base(name) ||
		strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid target list name %q", name)
	}
	return nil
}

// csvName appends .csv unless it's already there.
func csvName(name string) string {
	if !fileutil.HasExt(name, ".csv") {
		return name + ".csv"
	}
	return name
}

// GetTargetFileInfo returns metadata for a target list file.
// Accepts an absolute path, a filename ("mylist.csv"), or a bare name ("mylist").
func GetTargetFileInfo(name string, kind Kind) (FileInfo, error) {
	dir, err := baseDirForKind(kind)
	if err != nil {
		return FileInfo{}, err
	}

	fullPath := name
	if !filepath.IsAbs(name) {
		if err := validateListName(name); err != nil {
			return FileInfo{}, err
		}
		fullPath = filepath.Join(dir, csvName(name))
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return FileInfo{}, fmt.Errorf("%s list file not found: %s: %w", kind, name, ErrTargetFileNotFound)
		}
		return FileInfo{}, fmt.Errorf("stat file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return FileInfo{}, fmt.Errorf("not a regular file: %s", fullPath)
	}

	return FileInfo{
		Kind:      kind,
		Name:      fileutil.StripExt(filepath.Base(fullPath)),
		Path:      fullPath,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// ResolveTargetFilePath resolves a target list name to its absolute file path.
func ResolveTargetFilePath(name string, kind Kind) (string, error) {
	dir, err := baseDirForKind(kind)
	if err != nil {
		return "", err
	}

	if err := validateListName(name); err != nil {
		return "", err
	}

	return filepath.Join(dir, csvName(name)), nil
}

// TargetFileExists reports whether a target list file exists.
func TargetFileExists(name string, kind Kind) bool {
	path, err := ResolveTargetFilePath(name, kind)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
