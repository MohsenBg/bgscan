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
	"bgscan/internal/core/config/validate"
	"bgscan/internal/core/fileutil"
)

const (
	csvExtension    = ".csv"
	resultDirPerm   = 0o755
	timestampFormat = "20060102_150405"
)

// DefaultRegistry holds schemas used by GetResultFiles.
var DefaultRegistry = NewResultRegistry()

// baseDirOverride redirects the application base directory in tests.
// It is always empty in production.
var baseDirOverride string

// ResultRegistry stores result schemas safely for concurrent use.
type ResultRegistry struct {
	mu      sync.RWMutex
	schemas []ResultSchema
}

// NewResultRegistry returns an empty schema registry.
func NewResultRegistry() *ResultRegistry {
	return &ResultRegistry{}
}

// Register adds schema to the registry.
// It rejects invalid schemas and duplicate directories.
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

// Get returns the schema registered for directory.
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

// All returns a copy of the registered schemas, sorted by directory.
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

// Unregister removes the schema registered for directory.
// It reports whether a schema was removed.
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

// FindResultFiles returns metadata for CSV files in the given schema directories.
//
// Directories that do not exist are skipped. At least one schema is required.
func FindResultFiles(cfg config.WriterConfig, schemas ...ResultSchema) ([]ResultFile, error) {
	if len(schemas) == 0 {
		return nil, errors.New("result: at least one schema is required")
	}

	errs := validate.ValidateWriter(cfg)
	for _, err := range errs {
		return nil, err
	}

	var results []ResultFile

	for _, schema := range schemas {
		dir := getSchemaDir(cfg.ResultBaseDir, schema.Directory)

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

// GetResultFiles returns files for every schema in DefaultRegistry.
func GetResultFiles(cfg config.WriterConfig) ([]ResultFile, error) {
	schemas := DefaultRegistry.All()
	if len(schemas) == 0 {
		return nil, nil
	}
	return FindResultFiles(cfg, schemas...)
}

// ReadResultFile returns metadata for one result file.
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

// NormalizeResultFileName removes directory components and adds a .csv
// extension when one is missing.
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

// prepareResultFilePath builds a timestamped CSV path for schema and prefix.
// It creates the parent directory but does not create the file.
func prepareResultFilePath(cfg config.WriterConfig, schema ResultSchema, prefix string) (string, error) {
	if prefix == "" {
		return "", errors.New("result: prefix cannot be empty")
	}

	dir := getSchemaDir(cfg.ResultBaseDir, schema.Directory)

	if err := os.MkdirAll(dir, resultDirPerm); err != nil {
		return "", fmt.Errorf("create result directory %q: %w", dir, err)
	}

	filename := prefix + time.Now().Format(timestampFormat) + csvExtension
	return filepath.Join(dir, filename), nil
}

func getSchemaDir(resultDir, schemaDir string) string {
	base := baseDirOverride
	if base == "" {
		var err error
		if base, err = fileutil.BasePath(); err != nil {
			return filepath.Join(resultDir, schemaDir)
		}
	}

	return filepath.Join(base, resultDir, schemaDir)
}
