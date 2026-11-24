package result

import (
	"testing"
)

// TestResultSchema_Validate covers result schema validate.
func TestResultSchema_Validate(t *testing.T) {
	validParser := func([]string) (Result, error) {
		return nil, nil
	}

	tests := []struct {
		name    string
		schema  ResultSchema
		wantErr bool
	}{
		{
			name: "valid schema",
			schema: ResultSchema{
				Name:      "ports",
				Directory: "tcp",
				Parser:    validParser,
			},
		},
		{
			name: "missing name",
			schema: ResultSchema{
				Parser: validParser,
			},
			wantErr: true,
		},
		{
			name: "missing parser",
			schema: ResultSchema{
				Name: "ports",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestResultFile_SizeString covers result file size string.
func TestResultFile_SizeString(t *testing.T) {
	tests := []struct {
		sizeBytes int64
		want      string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{int64(2.5 * 1024 * 1024), "2.50 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{-1, "-1 B"},
	}

	for _, tt := range tests {
		file := ResultFile{SizeBytes: tt.sizeBytes}

		if got := file.SizeString(); got != tt.want {
			t.Errorf("SizeString() for %d bytes = %q, want %q",
				tt.sizeBytes, got, tt.want)
		}
	}
}

// TestMockResult_ImplementsResult covers mock result implements result.
func TestMockResult_ImplementsResult(t *testing.T) {
	var _ Result = newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9")
}
