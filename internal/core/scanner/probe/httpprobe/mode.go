package httpprobe

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Request is everything that can differ from one target to the next.
// Mode.Build creates one per entry; Settings never changes during a scan.
type Request struct {
	URL    string     // full URL to request
	Host   string     // Host header; empty means "use the URL host"
	SNI    string     // TLS server name; empty means "derive from the URL host"
	DialIP netip.Addr // valid: connect to this IP; invalid: normal DNS
	Port   uint16     // port actually probed (recorded in the result)
}

// Mode is the only place where "host mode" and "direct mode" differ: how a
// list entry turns into a Request.
type Mode interface {
	// Input is the kind of entry this mode accepts.
	Input() target.Kind

	// Build turns an entry into a Request, or rejects the entry.
	Build(e target.Entry) (Request, error)
}

// NewMode picks the mode once, at construction time.
func NewMode(cfg config.HTTPConfig, s Settings) (Mode, error) {
	parsed, err := config.ParseHTTPMode(string(cfg.Mode))
	if err != nil {
		return nil, err
	}

	switch parsed {
	case config.HTTPModeHost:
		return NewHostMode(cfg, s)
	case config.HTTPModeTarget:
		return NewDirectMode(cfg, s), nil
	default:
		return nil, fmt.Errorf("unsupported HTTP mode %q", parsed)
	}
}

func parseURL(cfg config.HTTPConfig, s Settings) (*url.URL, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("host is empty")
	}

	host, err := netutil.NormalizeHostWithSuffix(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("normalize host: %w", err)
	}

	u, err := url.Parse(s.scheme() + "://" + host)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	return u, nil
}

// HostMode probes one fixed host through many IPs (CDN scan). The entry only
// supplies the address to dial.
type HostMode struct {
	settings Settings
	base     url.URL
	host     string
	sni      string
}

// NewHostMode builds a HostMode from the configured fixed host.
func NewHostMode(cfg config.HTTPConfig, s Settings) (*HostMode, error) {
	host, err := netutil.ExtractTLSServerName(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("extract host: %w", err)
	}

	sni := cfg.ServerName
	if sni == "" {
		sni = host
	}

	base, err := parseURL(cfg, s)
	if err != nil {
		return nil, fmt.Errorf("parse host: %w", err)
	}

	return &HostMode{
		settings: s,
		host:     host,
		sni:      sni,
		base:     *base,
	}, nil
}

// Input returns target.KindIP. Host mode only takes addresses.
func (m *HostMode) Input() target.Kind { return target.KindIP }

// Build returns a Request that dials the entry address while asking
// for the fixed host. It rejects domain entries with
// probe.ErrDomainTarget.
func (m *HostMode) Build(e target.Entry) (Request, error) {
	if e.Kind != target.KindIP {
		return Request{}, fmt.Errorf("%w: %s", probe.ErrDomainTarget, e.Value())
	}

	port := m.settings.resolvePort(e)

	u := m.base
	u.Host = net.JoinHostPort(m.host, strconv.Itoa(int(port)))

	return Request{
		URL:    u.String(),
		Host:   m.host,
		SNI:    m.sni,
		DialIP: e.IP.Addr(),
		Port:   port,
	}, nil
}

// DirectMode probes each entry as it is: an IP is dialed directly, a domain
// is resolved by the normal dialer.
type DirectMode struct {
	settings Settings
	sni      string
}

// NewDirectMode builds a DirectMode. Only the path comes from config.
func NewDirectMode(cfg config.HTTPConfig, s Settings) *DirectMode {
	// The request path comes from each target entry, not from config.
	return &DirectMode{
		settings: s,
		sni:      cfg.ServerName,
	}
}

// Input returns target.KindHost. Direct mode takes an IP or a domain.
func (m *DirectMode) Input() target.Kind { return target.KindHost }

// Build returns a Request for the entry as written. An IP entry is
// pinned as the dial address. A domain entry has no dial address, so
// the normal dialer resolves it.
func (m *DirectMode) Build(e target.Entry) (Request, error) {
	var (
		host   string
		dialIP netip.Addr
	)

	switch e.Kind {
	case target.KindIP:
		dialIP = e.IP.Addr()
		host = dialIP.String()
	case target.KindDomain:
		host = e.Domain
	default:
		return Request{}, fmt.Errorf("unsupported target kind %s", e.Kind)
	}

	port := m.settings.resolvePort(e)
	u := url.URL{
		Scheme: m.settings.scheme(),
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   e.Path,
	}

	// Host and SNI stay empty: both are derived from the URL host.
	return Request{
		URL:    u.String(),
		DialIP: dialIP,
		Port:   port,
		SNI:    m.sni,
	}, nil
}
