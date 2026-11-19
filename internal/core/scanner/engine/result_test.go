package engine

import (
	"testing"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

func TestResultSchema_Validate(t *testing.T) {
	parser := func([]string) (result.Result, error) { return nil, nil }

	valid := result.ResultSchema{Name: "test", Parser: parser}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected no error for valid schema, got %v", err)
	}

	noName := result.ResultSchema{Parser: parser}
	if err := noName.Validate(); err == nil {
		t.Fatal("expected error for schema with no name")
	}

	noParser := result.ResultSchema{Name: "test"}
	if err := noParser.Validate(); err == nil {
		t.Fatal("expected error for schema with no parser")
	}
}

func TestResultFile_SizeString(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{int64(2.5 * 1024 * 1024 * 1024), "2.50 GB"},
	}
	for _, tc := range cases {
		f := result.ResultFile{SizeBytes: tc.bytes}
		if got := f.SizeString(); got != tc.want {
			t.Errorf("SizeString(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}
