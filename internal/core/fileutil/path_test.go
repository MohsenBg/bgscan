package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetCurrentPath_ReturnsAbsolutePath covers get current path returns absolute path.
func TestGetCurrentPath_ReturnsAbsolutePath(t *testing.T) {
	path, err := GetCurrentPath()
	if err != nil {
		t.Fatalf(
			"GetCurrentPath failed: %v",
			err,
		)
	}

	if !filepath.IsAbs(path) {
		t.Fatalf(
			"path should be absolute: %s",
			path,
		)
	}
}

// TestFileExists_ReturnsTrueForFile covers file exists returns true for file.
func TestFileExists_ReturnsTrueForFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"test.txt",
	)

	err := os.WriteFile(
		path,
		[]byte("hello"),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !FileExists(path) {
		t.Fatal(
			"expected file to exist",
		)
	}
}

// TestFileExists_ReturnsFalseForMissingFile covers file exists returns false for missing file.
func TestFileExists_ReturnsFalseForMissingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"missing.txt",
	)

	if FileExists(path) {
		t.Fatal(
			"expected missing file to return false",
		)
	}
}

// TestFileExists_ReturnsFalseForDirectory covers file exists returns false for directory.
func TestFileExists_ReturnsFalseForDirectory(t *testing.T) {
	dir := t.TempDir()

	if FileExists(dir) {
		t.Fatal(
			"directory should not count as file",
		)
	}
}

// TestStripExt_RemovesExtension covers strip ext removes extension.
func TestStripExt_RemovesExtension(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "config.json",
			want: "config",
		},
		{
			name: "archive.tar.gz",
			want: "archive.tar",
		},
		{
			name: "no-extension",
			want: "no-extension",
		},
		{
			name: "/tmp/file.txt",
			want: "/tmp/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripExt(tt.name)

			if got != tt.want {
				t.Fatalf(
					"StripExt(%q)=%q want %q",
					tt.name,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestHasExt_CaseInsensitive covers has ext case insensitive.
func TestHasExt_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		want bool
	}{
		{
			name: "config.JSON",
			ext:  ".json",
			want: true,
		},
		{
			name: "config.json",
			ext:  ".JSON",
			want: true,
		},
		{
			name: "config.txt",
			ext:  ".json",
			want: false,
		},
		{
			name: "config",
			ext:  ".json",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got := HasExt(
					tt.name,
					tt.ext,
				)

				if got != tt.want {
					t.Fatalf(
						"HasExt(%q,%q)=%v want %v",
						tt.name,
						tt.ext,
						got,
						tt.want,
					)
				}
			},
		)
	}
}

// TestEnsureParentDir_CreatesMissingParents covers ensure parent dir creates missing parents.
func TestEnsureParentDir_CreatesMissingParents(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"a",
		"b",
		"file.txt",
	)

	err := EnsureParentDir(path)
	if err != nil {
		t.Fatalf(
			"EnsureDir failed: %v",
			err,
		)
	}

	info, err := os.Stat(
		filepath.Dir(path),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !info.IsDir() {
		t.Fatal(
			"expected directory",
		)
	}
}

// TestEnsureParentDir_BareNameIsNoOp covers ensure parent dir bare name is no op.
func TestEnsureParentDir_BareNameIsNoOp(t *testing.T) {
	err := EnsureParentDir("file.txt")
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

// TestGetOrCreateBaseDir_CreatesMissingDirectory covers get or create base dir creates missing directory.
func TestGetOrCreateBaseDir_CreatesMissingDirectory(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"new",
		"directory",
	)

	got, err := GetOrCreateBaseDir(path)
	if err != nil {
		t.Fatalf(
			"GetOrCreateBaseDir failed: %v",
			err,
		)
	}

	if !filepath.IsAbs(got) {
		t.Fatalf(
			"expected absolute path got %s",
			got,
		)
	}

	info, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}

	if !info.IsDir() {
		t.Fatal(
			"expected directory",
		)
	}
}

// TestGetOrCreateBaseDir_ReturnsExistingDirectory covers get or create base dir returns existing directory.
func TestGetOrCreateBaseDir_ReturnsExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	got, err := GetOrCreateBaseDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Fatalf(
			"path mismatch\nwant=%s\ngot=%s",
			expected,
			got,
		)
	}
}

// TestGetOrCreateBaseDir_ReturnsErrorWhenPathIsFile covers get or create base dir returns error when path is file.
func TestGetOrCreateBaseDir_ReturnsErrorWhenPathIsFile(t *testing.T) {
	file := filepath.Join(
		t.TempDir(),
		"file.txt",
	)

	err := os.WriteFile(
		file,
		[]byte("data"),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetOrCreateBaseDir(file)

	if err == nil {
		t.Fatal(
			"expected error for file path",
		)
	}
}

// TestGetOrCreateBaseDir_ConvertsRelativePath covers get or create base dir converts relative path.
func TestGetOrCreateBaseDir_ConvertsRelativePath(t *testing.T) {
	base := t.TempDir()

	t.Chdir(base)

	got, err := GetOrCreateBaseDir(
		"relative-dir",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasSuffix(
		got,
		"relative-dir",
	) {
		t.Fatalf(
			"unexpected path: %s",
			got,
		)
	}

	if !filepath.IsAbs(got) {
		t.Fatal(
			"expected absolute path",
		)
	}
}

// TestRemoveFile_MissingIsSuccess covers remove file missing is success.
func TestRemoveFile_MissingIsSuccess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.csv")

	if err := RemoveFile(path); err != nil {
		t.Fatalf("RemoveFile on missing path: %v", err)
	}
}
