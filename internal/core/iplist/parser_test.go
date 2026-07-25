package iplist

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseRecord(t *testing.T) {
	tests := []struct {
		name       string
		input      []string
		wantIP     string
		wantEnable bool
		wantOk     bool
	}{
		{"ValidIP", []string{"1.1.1.1"}, "1.1.1.1", true, true},
		{"ValidIPWithEnable", []string{"1.1.1.1", "0"}, "1.1.1.1", false, true},
		{"CIDR", []string{"10.0.0.0/24", "1"}, "10.0.0.0/24", true, true},
		{"Spaces", []string{"  8.8.8.8  "}, "8.8.8.8", true, true},
		{"InvalidIP", []string{"not-an-ip"}, "", false, false},
		{"EmptyRecord", []string{}, "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseRecord(tt.input)
			if ok != tt.wantOk {
				t.Errorf("ParseRecord() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if ok {
				if got.IP != tt.wantIP || got.Enable != tt.wantEnable {
					t.Errorf("ParseRecord() = %v, want IP:%s Enable:%v", got, tt.wantIP, tt.wantEnable)
				}
			}
		})
	}
}

func TestStreamActiveIPsSequential(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stream_test.csv")

	// Prepare data: 1 active, 1 inactive, 1 active
	data := "1.1.1.1,1\n2.2.2.2,0\n3.3.3.3,1\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("RespectsLimit", func(t *testing.T) {
		out := make(chan string, 10)
		ctx := context.Background()

		err := StreamActiveIPs(ctx, path, 1, false, out)
		close(out)

		if err != nil && err.Error() != "EOF" { // EOF is expected from your logic when limit is hit
			t.Errorf("unexpected error: %v", err)
		}

		results := []string{}
		for ip := range out {
			results = append(results, ip)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 IP, got %v", results)
		}
		if results[0] != "1.1.1.1" {
			t.Errorf("expected 1.1.1.1, got %s", results[0])
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		out := make(chan string) // Unbuffered to force block
		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		err := streamActiveIPsSequential(ctx, path, 10, out)
		if err == nil {
			t.Error("expected context error, got nil")
		}
	})
}
