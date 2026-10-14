package xray

import (
	"encoding/json"
	"time"
)

// Protocol is an Xray outbound protocol name.
type Protocol string

const (
	ProtocolVLESS       Protocol = "vless"
	ProtocolVMess       Protocol = "vmess"
	ProtocolTrojan      Protocol = "trojan"
	ProtocolShadowsocks Protocol = "shadowsocks"
	ProtocolSocks       Protocol = "socks"
	ProtocolHTTP        Protocol = "http"
	ProtocolWireGuard   Protocol = "wireguard"
	ProtocolHysteria    Protocol = "hysteria"

	ProtocolFreedom   Protocol = "freedom"
	ProtocolBlackhole Protocol = "blackhole"
	ProtocolDNS       Protocol = "dns"
	ProtocolLoopback  Protocol = "loopback"
)

// Network is an Xray transport name.
type Network string

const (
	NetworkTCP         Network = "tcp"
	NetworkRaw         Network = "raw"
	NetworkKCP         Network = "kcp"
	NetworkWS          Network = "ws"
	NetworkHTTP        Network = "http"
	NetworkH2          Network = "h2"
	NetworkGRPC        Network = "grpc"
	NetworkHTTPUpgrade Network = "httpupgrade"
	NetworkXHTTP       Network = "xhttp"
	NetworkSplitHTTP   Network = "splithttp"
	NetworkQUIC        Network = "quic"
	NetworkDomainSock  Network = "domainsocket"
	NetworkHysteria    Network = "hysteria"
)

// Security is the transport security setting.
type Security string

const (
	SecurityNone    Security = "none"
	SecurityTLS     Security = "tls"
	SecurityReality Security = "reality"
	SecurityXTLS    Security = "xtls"
)

const (
	SocksAuthNone = "noauth"
	LogLevelNone  = "none"
)

// Canonical maps aliases and legacy names to their canonical value.
func (n Network) Canonical() Network {
	switch n {
	case "", NetworkRaw:
		return NetworkTCP
	case NetworkH2:
		return NetworkHTTP
	case NetworkSplitHTTP:
		return NetworkXHTTP
	default:
		return n
	}
}

// Config is the minimal scan config: one local inbound plus user outbounds.
type Config struct {
	Log       *LogConfig        `json:"log,omitempty"`
	Inbounds  []Inbound         `json:"inbounds"`
	Outbounds []json.RawMessage `json:"outbounds"`
}

type LogConfig struct {
	LogLevel string `json:"loglevel"`
}

// Inbound is the local SOCKS listener used for probing.
type Inbound struct {
	Tag      string          `json:"tag"`
	Listen   string          `json:"listen"`
	Port     uint16          `json:"port"`
	Protocol Protocol        `json:"protocol"`
	Settings SocksSettings   `json:"settings"`
	Sniffing *SniffingConfig `json:"sniffing,omitempty"`
}

type SocksSettings struct {
	Auth string `json:"auth"`
	UDP  bool   `json:"udp"`
	IP   string `json:"ip,omitempty"`
}

type SniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride,omitempty"`
	RouteOnly    bool     `json:"routeOnly,omitempty"`
}

// Outbound is one Xray exit config. Protocol-specific parts are kept raw, and
// any top-level key not modeled here (sendThrough, proxySettings, ...) is kept
// in Extra, so a round trip never loses fields.
type Outbound struct {
	Tag            string          `json:"tag,omitempty"`
	Protocol       Protocol        `json:"protocol"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	Mux            json.RawMessage `json:"mux,omitempty"`

	// Extra preserves unknown fields during round trips.
	Extra map[string]json.RawMessage `json:"-"`
}

var knownOutboundKeys = map[string]struct{}{
	"tag":            {},
	"protocol":       {},
	"settings":       {},
	"streamSettings": {},
	"mux":            {},
}

func (o *Outbound) UnmarshalJSON(data []byte) error {
	type plain Outbound

	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	extra, err := extraFields(data, knownOutboundKeys)
	if err != nil {
		return err
	}

	*o = Outbound(p)
	o.Extra = extra
	return nil
}

func (o Outbound) MarshalJSON() ([]byte, error) {
	type plain Outbound
	return marshalWithExtra(plain(o), o.Extra)
}

type StreamSettings struct {
	Network  Network  `json:"network,omitempty"`
	Security Security `json:"security,omitempty"`

	TLSSettings     *TLSSettings    `json:"tlsSettings,omitempty"`
	RealitySettings json.RawMessage `json:"realitySettings,omitempty"`

	TCPSettings         json.RawMessage `json:"tcpSettings,omitempty"`
	WSSettings          json.RawMessage `json:"wsSettings,omitempty"`
	GRPCSettings        json.RawMessage `json:"grpcSettings,omitempty"`
	HTTPUpgradeSettings json.RawMessage `json:"httpupgradeSettings,omitempty"`
	XHTTPSettings       json.RawMessage `json:"xhttpSettings,omitempty"`
	Sockopt             json.RawMessage `json:"sockopt,omitempty"`

	// Extra preserves transport settings not modeled above (kcpSettings,
	// quicSettings, hysteriaSettings, rawSettings, finalmask, ...).
	Extra map[string]json.RawMessage `json:"-"`
}

var knownStreamKeys = map[string]struct{}{
	"network":             {},
	"security":            {},
	"tlsSettings":         {},
	"realitySettings":     {},
	"tcpSettings":         {},
	"wsSettings":          {},
	"grpcSettings":        {},
	"httpupgradeSettings": {},
	"xhttpSettings":       {},
	"sockopt":             {},
}

func (s *StreamSettings) UnmarshalJSON(data []byte) error {
	type plain StreamSettings

	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	extra, err := extraFields(data, knownStreamKeys)
	if err != nil {
		return err
	}

	*s = StreamSettings(p)
	s.Extra = extra
	return nil
}

func (s StreamSettings) MarshalJSON() ([]byte, error) {
	type plain StreamSettings
	return marshalWithExtra(plain(s), s.Extra)
}

type TLSSettings struct {
	ServerName    string   `json:"serverName,omitempty"`
	ALPN          []string `json:"alpn,omitempty"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	AllowInsecure bool     `json:"allowInsecure,omitempty"`

	// Extra preserves unknown fields during round trips.
	Extra map[string]json.RawMessage `json:"-"`
}

var knownTLSKeys = map[string]struct{}{
	"serverName":    {},
	"alpn":          {},
	"fingerprint":   {},
	"allowInsecure": {},
}

func (t *TLSSettings) UnmarshalJSON(data []byte) error {
	type plain TLSSettings

	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	extra, err := extraFields(data, knownTLSKeys)
	if err != nil {
		return err
	}

	*t = TLSSettings(p)
	t.Extra = extra
	return nil
}

func (t TLSSettings) MarshalJSON() ([]byte, error) {
	type plain TLSSettings
	return marshalWithExtra(plain(t), t.Extra)
}

// Network returns the transport, defaulting to TCP like Xray does.
func (o Outbound) Network() Network {
	if ss := o.StreamSettings; ss != nil && ss.Network != "" {
		return ss.Network
	}
	return NetworkTCP
}

// Security returns the transport security, defaulting to none.
func (o Outbound) Security() Security {
	if ss := o.StreamSettings; ss != nil && ss.Security != "" {
		return ss.Security
	}
	return SecurityNone
}

func (o Outbound) IsProxy() bool {
	return isProxyProtocol(o.Protocol)
}

// normalizeWSALPN forces http/1.1 ALPN for WebSocket over TLS.
func (o *Outbound) normalizeWSALPN() {
	if o == nil {
		return
	}

	ss := o.StreamSettings
	if ss == nil || ss.TLSSettings == nil || ss.Network.Canonical() != NetworkWS {
		return
	}

	ss.TLSSettings.ALPN = []string{"http/1.1"}
}

// OutboundFile describes one outbound template file on disk.
type OutboundFile struct {
	Name        string
	Path        string
	CreatedTime time.Time
	Protocol    Protocol
	Network     Network
	Security    Security
}
