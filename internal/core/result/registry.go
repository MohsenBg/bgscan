package result

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/fileutil"
)

const (
	csvExtension    = ".csv"
	resultDirPerm   = 0o755
	timestampFormat = "20060102_150405"
)

// DefaultRegistry is the package-level registry used by GetResultFiles.
var DefaultRegistry = NewResultRegistry()

// ResultRegistry provides thread-safe storage for result schemas.
type ResultRegistry struct {
	mu      sync.RWMutex
	schemas []ResultSchema
}

// NewResultRegistry creates an empty ResultRegistry.
func NewResultRegistry() *ResultRegistry {
	return &ResultRegistry{}
}

// Register adds a schema to the registry. Returns an error if the schema
// is invalid or if a schema with the same Directory is already registered.
func (r *ResultRegistry) Register(schema ResultSchema) error {
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("register schema: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.schemas {
		if existing.Directory == schema.Directory {
			return fmt.Errorf("result: schema for directory %q already registered", schema.Directory)
		}
	}

	r.schemas = append(r.schemas, schema)
	return nil
}

// Get retrieves a schema by its Directory.
func (r *ResultRegistry) Get(directory string) (ResultSchema, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, schema := range r.schemas {
		if schema.Directory == directory {
			return schema, true
		}
	}
	return ResultSchema{}, false
}

// All returns all registered schemas, sorted by Directory for deterministic ordering.
func (r *ResultRegistry) All() []ResultSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]ResultSchema, len(r.schemas))
	copy(out, r.schemas)

	slices.SortFunc(out, func(a, b ResultSchema) int {
		return strings.Compare(a.Directory, b.Directory)
	})
	return out
}

// Len returns the number of registered schemas.
func (r *ResultRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.schemas)
}

// Unregister removes a schema from the registry by its Directory.
// Returns true if a schema was removed.
func (r *ResultRegistry) Unregister(directory string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, schema := range r.schemas {
		if schema.Directory == directory {
			r.schemas = slices.Delete(r.schemas, i, i+1)
			return true
		}
	}
	return false
}

// resultBaseDir returns the configured base directory for all result files.
func resultBaseDir() string {
	return config.GetWriter().ResultBaseDir
}

// FindResultFiles returns metadata for all result CSV files belonging to
// the provided schemas. At least one schema is required.
func FindResultFiles(schemas ...ResultSchema) ([]ResultFile, error) {
	if len(schemas) == 0 {
		return nil, errors.New("result: at least one schema is required")
	}

	baseDir := resultBaseDir()
	var results []ResultFile

	for _, schema := range schemas {
		dir := filepath.Join(baseDir, schema.Directory)

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.EqualFold(filepath.Ext(name), csvExtension) {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			results = append(results, ResultFile{
				Name:        fileutil.StripExt(name),
				SizeBytes:   info.Size(),
				CreatedTime: info.ModTime(),
				Path:        filepath.Join(dir, name),
				Schema:      schema,
				RecordCount: 0,
			})
		}
	}

	return results, nil
}

// GetResultFiles returns all result files across every schema registered
// in DefaultRegistry.
func GetResultFiles() ([]ResultFile, error) {
	schemas := DefaultRegistry.All()
	if len(schemas) == 0 {
		return nil, nil
	}
	return FindResultFiles(schemas...)
}

// ReadResultFile reads metadata for a single result file.
func ReadResultFile(path string, schema ResultSchema) (ResultFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ResultFile{}, fmt.Errorf("read result file %q: %w", path, err)
	}

	return ResultFile{
		Name:        fileutil.StripExt(info.Name()),
		SizeBytes:   info.Size(),
		CreatedTime: info.ModTime(),
		Path:        path,
		Schema:      schema,
		RecordCount: 0,
	}, nil
}

// NormalizeResultFileName ensures the name has a .csv extension and
// contains no directory components.
func NormalizeResultFileName(name string) string {
	if name == "" {
		return ".csv"
	}
	base := filepath.Base(name)
	if !strings.EqualFold(filepath.Ext(base), csvExtension) {
		return base + csvExtension
	}
	return base
}

// BuildResultFilePath creates a new output path using the provided schema
// and filename prefix. The base directory is read from config. The target
// directory is created if it does not already exist.
func BuildResultFilePath(schema ResultSchema, prefix string) (string, error) {
	if prefix == "" {
		return "", errors.New("result: prefix cannot be empty")
	}

	dir := filepath.Join(resultBaseDir(), schema.Directory)
	if err := os.MkdirAll(dir, resultDirPerm); err != nil {
		return "", fmt.Errorf("create result directory %q: %w", dir, err)
	}

	filename := prefix + time.Now().Format(timestampFormat) + csvExtension
	return filepath.Join(dir, filename), nil
}
