package xray

import "time"

const addressPlaceholder = "$ADDRESS"

// XrayConfig is the minimal scan config: local inbound + user outbounds.
type XrayConfig struct {
	// Inbound holds the proxy entry points (usually one local SOCKS).
	Inbound []Inbound `json:"inbounds"`

	// Outbound holds the exit configs. Kept as any since each proxy
	// protocol (VMess, VLESS, Trojan, ...) has its own schema.
	Outbound []any `json:"outbounds"`

	Log *xrayLogConfig `json:"log"`
}

type xrayLogConfig struct {
	Loglevel string `json:"loglevel"`
}

// Inbound is one proxy listener (usually localhost SOCKS for probes).
type Inbound struct {
	// Port is the TCP port the inbound listens on.
	Port uint16 `json:"port"`

	// Listen is the bind address (localhost only).
	Listen string `json:"listen"`

	// Tag identifies the inbound for routing rules.
	Tag string `json:"tag"`

	// Protocol is the inbound protocol (e.g. "socks").
	Protocol string `json:"protocol"`

	// Settings holds protocol-specific config.
	Settings SocksSettings `json:"settings"`

	// Sniffing enables protocol detection for routed traffic.
	Sniffing SniffingSetting `json:"sniffing"`
}

// SocksSettings is the config for a SOCKS inbound.
type SocksSettings struct {
	// Auth is the auth method ("noauth", local use only).
	Auth string `json:"auth"`

	// UDP enables UDP support.
	UDP bool `json:"udp"`

	// IP is the IP used for outbound UDP associations.
	IP string `json:"ip"`
}

// SniffingSetting controls protocol sniffing.
type SniffingSetting struct {
	// Enabled toggles sniffing.
	Enabled bool `json:"enabled"`

	// DestOverride lists protocols that override the destination address.
	DestOverride []string `json:"destOverride"`
}

type XrayOutboundsFile struct {
	Name        string    // File name without extension.
	CreatedTime time.Time // File modification time.
	Path        string    // Template file path.

	Protocol string // vless, vmess, trojan, shadowsocks...
	Network  string // tcp, ws, grpc, xhttp...
	UseTLS   bool   // whether transport security is TLS
}
