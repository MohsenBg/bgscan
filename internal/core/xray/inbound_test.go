package xray

import (
	"reflect"
	"testing"
)

func TestGetInbound(t *testing.T) {
	const port uint16 = 10808

	got := getInbound(port)

	want := Inbound{
		Port:     port,
		Listen:   "127.0.0.1",
		Tag:      "socks-inbound",
		Protocol: "socks",
		Settings: SocksSettings{
			Auth: "noauth",
			UDP:  false,
			IP:   "127.0.0.1",
		},
		Sniffing: SniffingSetting{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getInbound(%d) = %#v, want %#v", port, got, want)
	}
}

func TestGetInbound_PortVariants(t *testing.T) {
	tests := []struct {
		name string
		port uint16
	}{
		{"zero", 0},
		{"common socks port", 1080},
		{"high port", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInbound(tt.port)
			if got.Port != tt.port {
				t.Fatalf("getInbound(%d).Port = %d, want %d", tt.port, got.Port, tt.port)
			}
		})
	}
}
