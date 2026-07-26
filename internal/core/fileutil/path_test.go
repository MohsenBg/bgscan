package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestCheckFileExists_ReturnsTrueForFile(t *testing.T) {
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

	if !CheckFileExists(path) {
		t.Fatal(
			"expected file to exist",
		)
	}
}

func TestCheckFileExists_ReturnsFalseForMissingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"missing.txt",
	)

	if CheckFileExists(path) {
		t.Fatal(
			"expected missing file to return false",
		)
	}
}

func TestCheckFileExists_ReturnsFalseForDirectory(t *testing.T) {
	dir := t.TempDir()

	if CheckFileExists(dir) {
		t.Fatal(
			"directory should not count as file",
		)
	}
}

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

func TestEnsureDir_CreatesParentDirectories(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"a",
		"b",
		"file.txt",
	)

	err := EnsureDir(path)
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

func TestEnsureDir_NoParentDoesNothing(t *testing.T) {
	err := EnsureDir("file.txt")
	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}

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

func TestGetOrCreateBaseDir_ConvertsRelativePath(t *testing.T) {
	base := t.TempDir()

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(old)

	err = os.Chdir(base)
	if err != nil {
		t.Fatal(err)
	}

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
