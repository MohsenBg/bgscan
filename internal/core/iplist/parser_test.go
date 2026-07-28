package iplist

import (
	"context"
	"errors"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func mustAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	a, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("parse addr %q: %v", s, err)
	}
	return a
}

func TestParseRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      []string
		wantIP     netip.Prefix
		wantEnable bool
		wantOk     bool
	}{
		{
			name:       "ValidIP",
			input:      []string{"1.1.1.1"},
			wantIP:     mustPrefix(t, "1.1.1.1/32"),
			wantEnable: true,
			wantOk:     true,
		},
		{
			name:       "ValidIPWithEnable",
			input:      []string{"1.1.1.1", "0"},
			wantIP:     mustPrefix(t, "1.1.1.1/32"),
			wantEnable: false,
			wantOk:     true,
		},
		{
			name:       "CIDR",
			input:      []string{"10.0.0.0/24", "1"},
			wantIP:     mustPrefix(t, "10.0.0.0/24"),
			wantEnable: true,
			wantOk:     true,
		},
		{
			name:       "Spaces",
			input:      []string{"  8.8.8.8  "},
			wantIP:     mustPrefix(t, "8.8.8.8/32"),
			wantEnable: true,
			wantOk:     true,
		},
		{
			name:   "InvalidIP",
			input:  []string{"not-an-ip"},
			wantOk: false,
		},
		{
			name:   "EmptyRecord",
			input:  []string{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseRecord(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ParseRecord() ok = %v, want %v", ok, tt.wantOk)
			}
			if !ok {
				return
			}

			if got.IP != tt.wantIP || got.Enable != tt.wantEnable {
				t.Fatalf(
					"ParseRecord() = %+v, want IP:%v Enable:%v",
					got, tt.wantIP, tt.wantEnable,
				)
			}
		})
	}
}

func TestStreamActiveIPsSequential(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stream_test.csv")

	data := "" +
		"1.1.1.1,1\n" +
		"2.2.2.2,0\n" +
		"3.3.3.3,1\n"

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("RespectsLimit", func(t *testing.T) {
		t.Parallel()

		out := make(chan netip.Addr, 10)
		ctx := context.Background()

		err := StreamActiveIPs(ctx, path, 1, false, out)
		close(out)

		// depending on ReadCSV implementation, limit stop may surface as nil or io.EOF
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatalf("unexpected error: %v", err)
		}

		var results []netip.Addr
		for ip := range out {
			results = append(results, ip)
		}

		want := []netip.Addr{
			mustAddr(t, "1.1.1.1"),
		}

		if !reflect.DeepEqual(results, want) {
			t.Fatalf("results = %v, want %v", results, want)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		t.Parallel()

		out := make(chan netip.Addr) // unbuffered to force select path
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := streamActiveIPsSequential(ctx, path, 10, out)
		if err == nil {
			t.Fatal("expected context error, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}
