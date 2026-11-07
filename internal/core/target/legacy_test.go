package target

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyIPLists_CopiesMissingCSVFiles(t *testing.T) {
	base := setBaseDir(t)

	legacyDir := filepath.Join(base, LegacyIPListDir)
	mustWriteFile(t, filepath.Join(legacyDir, "legacy.csv"), "1.1.1.1,1\n")
	mustWriteFile(t, filepath.Join(legacyDir, "notes.txt"), "ignore me\n")

	// A canonical list with the same name must win.
	mustWriteFile(t, filepath.Join(base, IPListDir, "legacy.csv"), "9.9.9.9,1\n")
	mustWriteFile(t, filepath.Join(legacyDir, "extra.csv"), "2.2.2.2,1\n")

	copied, err := MigrateLegacyIPLists()
	if err != nil {
		t.Fatalf("MigrateLegacyIPLists() error = %v", err)
	}
	if copied != 1 {
		t.Fatalf("MigrateLegacyIPLists() copied = %d, want 1", copied)
	}

	got, err := os.ReadFile(filepath.Join(base, IPListDir, "legacy.csv"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "9.9.9.9,1\n" {
		t.Fatalf("existing target list was overwritten: %q", got)
	}

	got, err = os.ReadFile(filepath.Join(base, IPListDir, "extra.csv"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "2.2.2.2,1\n" {
		t.Fatalf("migrated list = %q, want source contents", got)
	}

	if _, err := os.Stat(filepath.Join(legacyDir, "extra.csv")); err != nil {
		t.Fatalf("legacy source was removed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(base, IPListDir, "notes.txt")); !os.IsNotExist(err) {
		t.Fatalf("non-CSV file was migrated: %v", err)
	}
}

func TestMigrateLegacyIPLists_MissingLegacyDir(t *testing.T) {
	setBaseDir(t)

	copied, err := MigrateLegacyIPLists()
	if err != nil {
		t.Fatalf("MigrateLegacyIPLists() error = %v", err)
	}
	if copied != 0 {
		t.Fatalf("MigrateLegacyIPLists() copied = %d, want 0", copied)
	}
}
