package target

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
		name    string
		input   []string
		want    Entry
		wantErr bool
	}{
		{
			name:  "ValidIP",
			input: []string{"1.1.1.1", "0", "", "1"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		},
		{
			name:  "ValidIPDisabled",
			input: []string{"1.1.1.1", "0", "", "0"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32")},
		},
		{
			name:  "CIDR",
			input: []string{"10.0.0.0/24", "0", "", "1"},
			want:  Entry{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		},
		{
			name:  "CIDRMasked",
			input: []string{"10.0.0.5/24", "0", "", "1"},
			want:  Entry{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		},
		{
			name:  "Spaces",
			input: []string{"  8.8.8.8  ", "0", "", "1"},
			want:  Entry{IP: mustPrefix(t, "8.8.8.8/32"), Enable: true},
		},
		{
			name:  "IPWithPort",
			input: []string{"8.8.8.8", "8080", "", "1"},
			want:  Entry{IP: mustPrefix(t, "8.8.8.8/32"), Port: 8080, Enable: true},
		},
		{
			name:  "IPWithPortAndPath",
			input: []string{"8.8.8.8", "443", "/resolve", "1"},
			want:  Entry{IP: mustPrefix(t, "8.8.8.8/32"), Port: 443, Path: "/resolve", Enable: true},
		},
		{
			name:  "CanonicalEmptyPortAndPath",
			input: []string{"8.8.8.8", "0", "", "1"},
			want:  Entry{IP: mustPrefix(t, "8.8.8.8/32"), Enable: true},
		},
		{
			name:  "Domain",
			input: []string{"example.com", "0", "", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Enable: true},
		},
		{
			name:  "DomainLowercased",
			input: []string{"EXAMPLE.com", "0", "", "0"},
			want:  Entry{Kind: KindDomain, Domain: "example.com"},
		},
		{
			name:  "DomainWithPort",
			input: []string{"example.com", "8443", "", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Port: 8443, Enable: true},
		},
		{
			name:  "DomainWithPortAndPath",
			input: []string{"example.com", "443", "/dns-query", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/dns-query", Enable: true},
		},
		{
			name:  "DomainWithEmptyPort",
			input: []string{"example.com", "", "/path", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Path: "/path", Enable: true},
		},
		{
			name:  "PathWithoutLeadingSlashIsNormalized",
			input: []string{"example.com", "443", "dns-query", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/dns-query", Enable: true},
		},
		{
			name:  "TrailingMeasurementsIgnored",
			input: []string{"1.1.1.1", "443", "/x", "1", "extra", "more"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32"), Port: 443, Path: "/x", Enable: true},
		},
		{
			name:    "ResultRowRejected",
			input:   []string{"1.1.1.1", "12ms", "443", "1", "3"},
			wantErr: true,
		},
		{
			name:    "LegacyOneColumnRejected",
			input:   []string{"1.1.1.1"},
			wantErr: true,
		},
		{
			name:    "LegacyTwoColumnRejected",
			input:   []string{"1.1.1.1", "1"},
			wantErr: true,
		},
		{
			name:    "LegacyThreeColumnRejected",
			input:   []string{"1.1.1.1", "8080", "1"},
			wantErr: true,
		},
		{
			name:    "EmptyRecord",
			input:   []string{},
			wantErr: true,
		},
		{
			name:    "EmptyValue",
			input:   []string{"", "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "InvalidIP",
			input:   []string{"not-an-ip", "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "SchemeRejected",
			input:   []string{"https://example.com", "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "PathInValueRejected",
			input:   []string{"example.com/path", "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "SingleLabelRejected",
			input:   []string{"localhost", "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "InvalidPort",
			input:   []string{"example.com", "70000", "", "1"},
			wantErr: true,
		},
		{
			name:    "NonNumericPort",
			input:   []string{"example.com", "abc", "", "1"},
			wantErr: true,
		},
		{
			name:    "PathQueryRejected",
			input:   []string{"example.com", "443", "/x?y", "1"},
			wantErr: true,
		},
		{
			name:    "PathSpaceRejected",
			input:   []string{"example.com", "443", "/x y", "1"},
			wantErr: true,
		},
		{
			name:    "InvalidEnable",
			input:   []string{"example.com", "443", "", "yes"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRecord(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseRecord() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRecord() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseRecord() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseInputRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  Entry
	}{
		{
			name:  "OneColumnDefaultsEnabled",
			input: []string{"1.1.1.1"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		},
		{
			name:  "TwoColumnsEnable",
			input: []string{"1.1.1.1", "0"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32")},
		},
		{
			name:  "ThreeColumnsPortEnable",
			input: []string{"1.1.1.1", "8080", "1"},
			want:  Entry{IP: mustPrefix(t, "1.1.1.1/32"), Port: 8080, Enable: true},
		},
		{
			name:  "FourColumns",
			input: []string{"example.com", "443", "/x", "1"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/x", Enable: true},
		},
		{
			name:  "DomainOneColumn",
			input: []string{"example.com"},
			want:  Entry{Kind: KindDomain, Domain: "example.com", Enable: true},
		},
		{
			name:  "CIDROneColumn",
			input: []string{"10.0.0.5/24"},
			want:  Entry{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseInputRecord(tt.input)
			if err != nil {
				t.Fatalf("ParseInputRecord() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseInputRecord() = %#v, want %#v", got, tt.want)
			}
		})
	}

	if _, err := ParseInputRecord([]string{"a", "b", "c", "d", "e"}); err == nil {
		t.Fatal("ParseInputRecord() accepted five fields")
	}
}

func TestStreamActiveSequential(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stream_test.csv")

	data := "" +
		"1.1.1.1,0,,1\n" +
		"2.2.2.2,0,,0\n" +
		"3.3.3.3,0,,1\n"

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("RespectsLimit", func(t *testing.T) {
		t.Parallel()

		out := make(chan netip.Addr, 10)
		ctx := context.Background()

		err := StreamActive(nil, ctx, path, 1, false, out)
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

		err := streamActiveSequential(nil, ctx, path, 10, out)
		if err == nil {
			t.Fatal("expected context error, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})
}

func TestStreamActiveSequential_Guards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr bool
		wantIPs []string
	}{
		{
			name:    "PortIsIgnoredWithWarning",
			content: "8.8.8.8,8080,,1\n",
			wantIPs: []string{"8.8.8.8"},
		},
		{
			name:    "PathRejected",
			content: "8.8.8.8,443,/resolve,1\n",
			wantErr: true,
		},
		{
			name:    "DomainRejected",
			content: "example.com,0,,1\n",
			wantErr: true,
		},
		{
			name:    "DisabledDomainSkipped",
			content: "example.com,0,,0\n",
			wantIPs: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "guard.csv")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			out := make(chan netip.Addr, 10)
			err := streamActiveSequential(nil, context.Background(), path, 0, out)
			close(out)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var got []string
			for ip := range out {
				got = append(got, ip.String())
			}
			if !reflect.DeepEqual(got, tt.wantIPs) {
				t.Fatalf("streamed IPs = %v, want %v", got, tt.wantIPs)
			}
		})
	}
}
