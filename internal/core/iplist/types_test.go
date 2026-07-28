package iplist

import (
	"net/netip"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		ip     netip.Prefix
		enable int
		want   IPList
	}{
		{
			name:   "EnabledIPv4Single",
			ip:     mustPrefix(t, "1.1.1.1"),
			enable: 1,
			want:   IPList{IP: mustPrefix(t, "1.1.1.1"), Enable: true},
		},
		{
			name:   "DisabledIPv4Single",
			ip:     mustPrefix(t, "1.1.1.1"),
			enable: 0,
			want:   IPList{IP: mustPrefix(t, "1.1.1.1"), Enable: false},
		},
		{
			name:   "InvalidEnableValue",
			ip:     mustPrefix(t, "1.1.1.1"),
			enable: 2,
			want:   IPList{IP: mustPrefix(t, "1.1.1.1"), Enable: false},
		},
		{
			name:   "EnabledCIDR",
			ip:     mustPrefix(t, "10.0.0.0/24"),
			enable: 1,
			want:   IPList{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		},
		{
			name:   "EnabledIPv6Single",
			ip:     mustPrefix(t, "2001:db8::1"),
			enable: 1,
			want:   IPList{IP: mustPrefix(t, "2001:db8::1"), Enable: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.ip, tt.enable)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNewEnabledAndNewDisabled(t *testing.T) {
	tests := []struct {
		name string
		ip   netip.Prefix
		want IPList
		fn   func(netip.Prefix) IPList
	}{
		{
			name: "NewEnabledIPv4",
			ip:   mustPrefix(t, "1.1.1.1"),
			want: IPList{IP: mustPrefix(t, "1.1.1.1"), Enable: true},
			fn:   NewEnabled,
		},
		{
			name: "NewDisabledIPv4",
			ip:   mustPrefix(t, "1.1.1.1"),
			want: IPList{IP: mustPrefix(t, "1.1.1.1"), Enable: false},
			fn:   NewDisabled,
		},
		{
			name: "NewEnabledCIDR",
			ip:   mustPrefix(t, "10.0.0.0/24"),
			want: IPList{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
			fn:   NewEnabled,
		},
		{
			name: "NewDisabledIPv6",
			ip:   mustPrefix(t, "2001:db8::1"),
			want: IPList{IP: mustPrefix(t, "2001:db8::1"), Enable: false},
			fn:   NewDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.ip)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("constructor() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestIPList_IsSingle(t *testing.T) {
	tests := []struct {
		name string
		ip   netip.Prefix
		want bool
	}{
		{
			name: "IPv4Single",
			ip:   mustPrefix(t, "192.168.1.1"),
			want: true,
		},
		{
			name: "IPv6Single",
			ip:   mustPrefix(t, "2001:db8::1"),
			want: true,
		},
		{
			name: "IPv4CIDR24",
			ip:   mustPrefix(t, "10.0.0.0/24"),
			want: false,
		},
		{
			name: "IPv4CIDR8",
			ip:   mustPrefix(t, "10.0.0.0/8"),
			want: false,
		},
		{
			name: "IPv6CIDR64",
			ip:   mustPrefix(t, "2001:db8::/64"),
			want: false,
		},
		{
			name: "IPv4HostPrefix32",
			ip:   mustPrefix(t, "8.8.8.8/32"),
			want: true,
		},
		{
			name: "IPv6HostPrefix128",
			ip:   mustPrefix(t, "2001:db8::1/128"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := IPList{IP: tt.ip}
			if got := i.IsSingle(); got != tt.want {
				t.Errorf("IPList.IsSingle(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIPList_EncodeCSV(t *testing.T) {
	tests := []struct {
		name string
		in   IPList
		want []string
	}{
		{
			name: "EnabledSingleIPv4",
			in:   IPList{IP: mustPrefix(t, "1.2.3.4"), Enable: true},
			want: []string{"1.2.3.4/32", "1"},
		},
		{
			name: "DisabledSingleIPv4",
			in:   IPList{IP: mustPrefix(t, "1.2.3.4"), Enable: false},
			want: []string{"1.2.3.4/32", "0"},
		},
		{
			name: "EnabledCIDR",
			in:   IPList{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
			want: []string{"10.0.0.0/24", "1"},
		},
		{
			name: "EnabledSingleIPv6",
			in:   IPList{IP: mustPrefix(t, "2001:db8::1"), Enable: true},
			want: []string{"2001:db8::1/128", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.EncodeCSV()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
