package netutil

import (
	"context"
	"testing"
	"time"
)

func TestStreamCIDR(t *testing.T) {
	ctx := context.Background()

	// Test normal IPv4 iteration
	out := make(chan string, 10)
	err := StreamCIDR(ctx, "192.168.1.0/30", 0, out)
	if err != nil {
		t.Fatalf("StreamCIDR() error = %v", err)
	}
	close(out)

	var ips []string
	for ip := range out {
		ips = append(ips, ip)
	}

	expected := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}
	if len(ips) != len(expected) {
		t.Fatalf("expected %d ips, got %d", len(expected), len(ips))
	}
	for i, ip := range ips {
		if ip != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], ip)
		}
	}

	// Test limit
	out2 := make(chan string, 10)
	err = StreamCIDR(ctx, "192.168.1.0/30", 2, out2)
	if err != nil {
		t.Fatalf("StreamCIDR() with limit error = %v", err)
	}
	close(out2)
	ips2 := []string{}
	for ip := range out2 {
		ips2 = append(ips2, ip)
	}
	if len(ips2) != 2 {
		t.Errorf("expected 2 ips with limit, got %d", len(ips2))
	}

	// Test IPv6 with limit
	out4 := make(chan string, 10)
	err = StreamCIDR(ctx, "2001:db8::/126", 2, out4)
	if err != nil {
		t.Fatalf("StreamCIDR() IPv6 error = %v", err)
	}
	close(out4)
	ips4 := []string{}
	for ip := range out4 {
		ips4 = append(ips4, ip)
	}
	if len(ips4) != 2 {
		t.Errorf("expected 2 ips for IPv6 with limit, got %d", len(ips4))
	}

	// Test context cancellation
	ctxCancel, cancel := context.WithCancel(context.Background())
	out3 := make(chan string) // Unbuffered to force blocking

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err = StreamCIDR(ctxCancel, "10.0.0.0/8", 0, out3)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Test invalid CIDR
	err = StreamCIDR(ctx, "invalid", 0, out)
	if err == nil {
		t.Errorf("expected error for invalid CIDR")
	}
}

func TestParseIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{" 192.168.1.1 ", "192.168.1.1"},
		{"192.168.1.0/24", ""}, // CIDR rejected
		{"invalid", ""},
	}
	for _, tt := range tests {
		got := ParseIP(tt.input)
		if tt.want == "" {
			if got != nil {
				t.Errorf("ParseIP(%q) = %v, want nil", tt.input, got)
			}
		} else {
			if got == nil || got.String() != tt.want {
				t.Errorf("ParseIP(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestParseCIDR(t *testing.T) {
	_, err := ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Errorf("ParseCIDR() valid error = %v", err)
	}

	_, err = ParseCIDR("invalid")
	if err == nil {
		t.Errorf("ParseCIDR() expected error for invalid CIDR")
	}
}

func TestNormalizeIPOrCIDR(t *testing.T) {
	tests := []struct {
		raw    string
		want   string
		wantOk bool
	}{
		{"192.168.001.001", "192.168.1.1", true},
		{"10.0.0.0/8", "10.0.0.0/8", true},
		{" 10.0.0.0/8 ", "10.0.0.0/8", true},
		{"invalid", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalizeIPOrCIDR(tt.raw)
		if ok != tt.wantOk || got != tt.want {
			t.Errorf("NormalizeIPOrCIDR(%q) = %v, %v, want %v, %v", tt.raw, got, ok, tt.want, tt.wantOk)
		}
	}
}
