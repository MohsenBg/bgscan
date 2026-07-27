package netutil

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

func TestStreamCIDR(t *testing.T) {
	ctx := context.Background()

	// Test normal IPv4 iteration
	out := make(chan netip.Addr, 10)
	err := StreamCIDR(ctx, "192.168.1.0/30", 0, out)
	if err != nil {
		t.Fatalf("StreamCIDR() error = %v", err)
	}
	close(out)

	var ips []netip.Addr
	for ip := range out {
		ips = append(ips, ip)
	}

	expected := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}
	if len(ips) != len(expected) {
		t.Fatalf("expected %d ips, got %d", len(expected), len(ips))
	}
	for i, ip := range ips {
		if ip.String() != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], ip.String())
		}
	}

	// Test limit
	out2 := make(chan netip.Addr, 10)
	err = StreamCIDR(ctx, "192.168.1.0/30", 2, out2)
	if err != nil {
		t.Fatalf("StreamCIDR() with limit error = %v", err)
	}
	close(out2)
	count2 := 0
	for range out2 {
		count2++
	}
	if count2 != 2 {
		t.Errorf("expected 2 ips with limit, got %d", count2)
	}

	// Test IPv6 with limit
	out4 := make(chan netip.Addr, 10)
	err = StreamCIDR(ctx, "2001:db8::/126", 2, out4)
	if err != nil {
		t.Fatalf("StreamCIDR() IPv6 error = %v", err)
	}
	close(out4)
	count4 := 0
	for range out4 {
		count4++
	}
	if count4 != 2 {
		t.Errorf("expected 2 ips for IPv6 with limit, got %d", count4)
	}

	// Test context cancellation
	ctxCancel, cancel := context.WithCancel(context.Background())
	out3 := make(chan netip.Addr) // Unbuffered to ensure blocking

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

func TestParseIPOrCIDR(t *testing.T) {
	tests := []struct {
		raw    string
		want   string
		wantOk bool
	}{
		// Plain IP becomes /32
		{"192.168.1.1", "192.168.1.1/32", true},
		// CIDR remains CIDR
		{"10.0.0.0/8", "10.0.0.0/8", true},
		// Spacing
		{" 10.0.0.0/8 ", "10.0.0.0/8", true},
		// IPv6 normalization
		{"2001:db8::01", "2001:db8::1/128", true},
		// Invalid
		{"invalid", "", false},
	}

	for _, tt := range tests {
		got, ok := ParseIPOrCIDR(tt.raw)
		if ok != tt.wantOk {
			t.Errorf("ParseIPOrCIDR(%q) ok = %v, want %v", tt.raw, ok, tt.wantOk)
			continue
		}
		if ok && got.String() != tt.want {
			t.Errorf("ParseIPOrCIDR(%q) = %s, want %s", tt.raw, got.String(), tt.want)
		}
	}
}
