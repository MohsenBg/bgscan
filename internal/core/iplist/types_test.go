package iplist

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		enable int
		want   IPList
	}{
		{"Enabled", "1.1.1.1", 1, IPList{IP: "1.1.1.1", Enable: true}},
		{"Disabled", "1.1.1.1", 0, IPList{IP: "1.1.1.1", Enable: false}},
		{"InvalidEnableValue", "1.1.1.1", 2, IPList{IP: "1.1.1.1", Enable: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.ip, tt.enable); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.enable == 1 {
				if got := NewEnabled(tt.ip); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewEnabled() = %v, want %v", got, tt.want)
				}
			} else {
				if got := NewDisabled(tt.ip); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewEnabled() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIPList_IsCIDR(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"SimpleIPv4", "192.168.1.1", false},
		{"SimpleIPv6", "2001:db8::1", false}, // Note: your current logic checks for digits at start/end, IPv6 might fail this logic
		{"IPv4CIDR_24", "10.0.0.0/24", true},
		{"IPv4CIDR_8", "10.0.0.0/8", true},
		{"EdgeCase_Invalid", "1.1.1.1/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := IPList{IP: tt.ip}
			if got := i.IsCIDR(); got != tt.want {
				t.Errorf("IPList.IsCIDR(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIPList_EncodeCSV(t *testing.T) {
	i := IPList{IP: "1.2.3.4", Enable: true}
	got := i.EncodeCSV()
	want := []string{"1.2.3.4", "1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EncodeCSV() = %v, want %v", got, want)
	}
}
