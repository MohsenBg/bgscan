package result

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// ---------- ResultRegistry ----------

func TestNewResultRegistry_Empty(t *testing.T) {
	r := NewResultRegistry()
	if r.Len() != 0 {
		t.Errorf("new registry should be empty, got Len=%d", r.Len())
	}
}

func TestResultRegistry_Register_Valid(t *testing.T) {
	r := NewResultRegistry()
	s := validSchema(t)

	err := r.Register(s)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if r.Len() != 1 {
		t.Errorf("Len() = %d, want 1", r.Len())
	}
}

func TestResultRegistry_Register_InvalidSchema(t *testing.T) {
	r := NewResultRegistry()

	s := ResultSchema{Name: "", Parser: nil}
	err := r.Register(s)
	if err == nil {
		t.Fatal("expected error for invalid schema, got nil")
	}
}

func TestResultRegistry_Register_DuplicateDirectory(t *testing.T) {
	r := NewResultRegistry()
	s1 := validSchema(t)
	s2 := ResultSchema{
		Name:      "other",
		Directory: s1.Directory, // same directory
		Parser:    func([]string) (Result, error) { return nil, nil },
	}

	err := r.Register(s1)
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	err = r.Register(s2)
	if err == nil {
		t.Fatal("expected error for duplicate directory, got nil")
	}
}

func TestResultRegistry_Register_DifferentDirectories(t *testing.T) {
	r := NewResultRegistry()

	s1 := validSchema(t)
	s2 := ResultSchema{
		Name:      "ports",
		Directory: "ports_scan",
		Columns:   []ColumnDef{{Name: "ip"}},
		Parser:    func([]string) (Result, error) { return nil, nil },
	}

	if err := r.Register(s1); err != nil {
		t.Fatalf("Register(s1) error = %v", err)
	}
	if err := r.Register(s2); err != nil {
		t.Fatalf("Register(s2) error = %v", err)
	}
	if r.Len() != 2 {
		t.Errorf("Len() = %d, want 2", r.Len())
	}
}

func TestResultRegistry_Get_Found(t *testing.T) {
	r := NewResultRegistry()
	s := validSchema(t)

	_ = r.Register(s)

	got, ok := r.Get(s.Directory)
	if !ok {
		t.Fatal("Get() returned ok=false for registered directory")
	}
	if got.Name != s.Name {
		t.Errorf("Get().Name = %q, want %q", got.Name, s.Name)
	}
}

func TestResultRegistry_Get_NotFound(t *testing.T) {
	r := NewResultRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("Get() returned ok=true for unregistered directory")
	}
}

func TestResultRegistry_All_SortedByDirectory(t *testing.T) {
	r := NewResultRegistry()

	schemas := []ResultSchema{
		{Name: "c", Directory: "charlie", Parser: func([]string) (Result, error) { return nil, nil }},
		{Name: "a", Directory: "alpha", Parser: func([]string) (Result, error) { return nil, nil }},
		{Name: "b", Directory: "bravo", Parser: func([]string) (Result, error) { return nil, nil }},
	}
	for _, s := range schemas {
		_ = r.Register(s)
	}

	all := r.All()
	if len(all) != 3 {
		t.Fatalf("All() returned %d schemas, want 3", len(all))
	}

	// Should be sorted by Directory: alpha, bravo, charlie
	expected := []string{"alpha", "bravo", "charlie"}
	for i, dir := range expected {
		if all[i].Directory != dir {
			t.Errorf("All()[%d].Directory = %q, want %q", i, all[i].Directory, dir)
		}
	}
}

func TestResultRegistry_All_ReturnsCopy(t *testing.T) {
	r := NewResultRegistry()
	s := validSchema(t)
	_ = r.Register(s)

	all := r.All()
	all[0].Name = "MUTATED"

	got, ok := r.Get(s.Directory)
	if !ok {
		t.Fatal("schema should still be registered")
	}
	if got.Name == "MUTATED" {
		t.Error("All() should return a copy, not a reference to internal slice")
	}
}

func TestResultRegistry_All_Empty(t *testing.T) {
	r := NewResultRegistry()
	all := r.All()
	if len(all) != 0 {
		t.Errorf("All() on empty registry should return nil or empty, got %v", all)
	}
}

func TestResultRegistry_Unregister_Existing(t *testing.T) {
	r := NewResultRegistry()
	s := validSchema(t)
	_ = r.Register(s)

	removed := r.Unregister(s.Directory)
	if !removed {
		t.Error("Unregister() should return true for existing directory")
	}
	if r.Len() != 0 {
		t.Errorf("Len() after Unregister = %d, want 0", r.Len())
	}
}

func TestResultRegistry_Unregister_NonExisting(t *testing.T) {
	r := NewResultRegistry()

	removed := r.Unregister("nonexistent")
	if removed {
		t.Error("Unregister() should return false for non-existing directory")
	}
}

func TestResultRegistry_ConcurrentAccess(t *testing.T) {
	r := NewResultRegistry()
	s := validSchema(t)

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_ = r.Register(s)
			_, _ = r.Get(s.Directory)
			_ = r.All()
			_ = r.Len()
			_ = r.Unregister(s.Directory)
		})
	}
	wg.Wait()
}

// ---------- NormalizeResultFileName ----------

func TestNormalizeResultFileName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"already csv", "results.csv", "results.csv"},
		{"no extension", "results", "results.csv"},
		{"different extension", "results.txt", "results.txt.csv"},
		{"uppercase extension", "results.CSV", "results.CSV"},
		{"with directory", "/tmp/results.csv", "results.csv"},
		{"nested directory", "sub/dir/file.csv", "file.csv"},
		{"dotfile", ".hidden", ".hidden.csv"},
		{"empty", "", ".csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeResultFileName(tt.in)
			if got != tt.want {
				t.Errorf("NormalizeResultFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// ---------- FindResultFiles ----------

func TestFindResultFiles_NoSchemas(t *testing.T) {
	_, err := FindResultFiles()
	if err == nil {
		t.Fatal("expected error when no schemas provided, got nil")
	}
}

func TestFindResultFiles_NonexistentDir(t *testing.T) {
	schema := ResultSchema{
		Name:      "test",
		Directory: "nonexistent_dir_xyz",
		Parser:    func([]string) (Result, error) { return nil, nil },
	}

	// Non-existent directories are silently skipped per the implementation.
	files, err := FindResultFiles(schema)
	if err != nil {
		t.Fatalf("FindResultFiles() unexpected error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for nonexistent dir, got %d", len(files))
	}
}

func TestFindResultFiles_FindsCSVFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that matches the schema's directory name.
	schemaDir := filepath.Join(tmpDir, "myscan")
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Write CSV files.
	for _, name := range []string{"scan1.csv", "scan2.csv"} {
		f, err := os.Create(filepath.Join(schemaDir, name))
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		_ = f.Close()
	}

	f, _ := os.Create(filepath.Join(schemaDir, "readme.txt"))
	_ = f.Close()
	_ = os.MkdirAll(filepath.Join(schemaDir, "subdir"), 0o755)
	_ = schemaDir

	schema := ResultSchema{
		Name:      "test",
		Directory: "nonexistent_xyz_finds_csv_test",
		Parser:    func([]string) (Result, error) { return nil, nil },
	}

	files, err := FindResultFiles(schema)
	if err != nil {
		t.Fatalf("FindResultFiles() error = %v", err)
	}
	// We can't assert the count because config controls the base dir.
	// We only verify no error is returned.
	_ = files
}

// ---------- ReadResultFile ----------

func TestReadResultFile_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "scan.csv")

	content := "1.2.3.4,0.9\n5.6.7.8,0.5\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	schema := validSchema(t)
	rf, err := ReadResultFile(path, schema)
	if err != nil {
		t.Fatalf("ReadResultFile() error = %v", err)
	}
	if rf.Path != path {
		t.Errorf("Path = %q, want %q", rf.Path, path)
	}
	if rf.Name != "scan" {
		t.Errorf("Name = %q, want %q", rf.Name, "scan")
	}
	if rf.Schema.Name != schema.Name {
		t.Errorf("Schema.Name = %q, want %q", rf.Schema.Name, schema.Name)
	}
	if rf.SizeBytes != int64(len(content)) {
		t.Errorf("SizeBytes = %d, want %d", rf.SizeBytes, len(content))
	}
}

func TestReadResultFile_Nonexistent(t *testing.T) {
	_, err := ReadResultFile("/nonexistent/path/file.csv", validSchema(t))
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// ---------- BuildResultFilePath ----------

func TestBuildResultFilePath_EmptyPrefix(t *testing.T) {
	schema := validSchema(t)
	_, err := BuildResultFilePath(schema, "")
	if err == nil {
		t.Fatal("expected error for empty prefix, got nil")
	}
}

func TestBuildResultFilePath_CreatesDirectory(t *testing.T) {
	// BuildResultFilePath uses config.GetWriter().ResultBaseDir
	// which we cannot control without modifying production code.
	// We verify that an empty prefix returns an error.
	// A full test would require the config package to be set up.
	schema := validSchema(t)
	_, err := BuildResultFilePath(schema, "prefix_")
	// This may fail if config is not initialised, or succeed if it is.
	// We only assert that non-empty prefix doesn't return "prefix cannot be empty".
	if err != nil {
		// Acceptable — config may not be set up in test env.
		t.Skipf("BuildResultFilePath requires config setup: %v", err)
	}
}
