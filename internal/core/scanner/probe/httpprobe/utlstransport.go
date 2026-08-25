package httpprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"

	"bgscan/internal/core/config"
)

// utlsDialContext connects to addr using uTLS and forces the connection
// to targetIP while preserving the original port.
func utlsDialContext(
	ctx context.Context,
	network, addr string,
	targetIP netip.Addr,
	utlsConfig *utls.Config,
	clientHelloID *utls.ClientHelloID,
) (*utls.UConn, error) {
	if utlsConfig == nil {
		utlsConfig = &utls.Config{}
	}

	if utlsConfig.ServerName == "" {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("parse address: %w", err)
		}

		utlsConfig = utlsConfig.Clone()
		utlsConfig.ServerName = host
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("parse address: %w", err)
	}

	dialAddr := net.JoinHostPort(targetIP.String(), port)

	conn, err := (&net.Dialer{}).DialContext(ctx, network, dialAddr)
	if err != nil {
		return nil, err
	}

	uconn := utls.UClient(conn, utlsConfig, *clientHelloID)

	if err := uconn.HandshakeContext(ctx); err != nil {
		_ = uconn.Close()
		return nil, err
	}

	return uconn, nil
}

// utlsRoundTripper is an http.RoundTripper that uses uTLS for HTTPS
// connections.
//
// The first connection is used to determine the negotiated ALPN protocol.
// Once the underlying transport is created, its cleanup function is stored
// so the caller can release idle connections when the probe finishes.
type utlsRoundTripper struct {
	clientHelloID *utls.ClientHelloID
	config        *utls.Config
	targetIP      netip.Addr
	version       HTTPVersion

	mu    sync.Mutex
	inner http.RoundTripper
	close func()
}

func newUTLSRoundTripper(
	targetIP netip.Addr,
	utlsConfig *utls.Config,
	clientHelloID *utls.ClientHelloID,
	version HTTPVersion,
) *utlsRoundTripper {
	return &utlsRoundTripper{
		clientHelloID: clientHelloID,
		config:        utlsConfig,
		targetIP:      targetIP,
		version:       version,
	}
}

func (rt *utlsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Scheme {
	case "http":
		return http.DefaultTransport.RoundTrip(req)

	case "https":
	default:
		return nil, fmt.Errorf("unsupported URL scheme %q", req.URL.Scheme)
	}

	rt.mu.Lock()

	if rt.inner == nil {
		inner, closeFunc, err := rt.makeRoundTripper(req)
		if err != nil {
			rt.mu.Unlock()
			return nil, err
		}

		rt.inner = inner
		rt.close = closeFunc
	}

	inner := rt.inner
	rt.mu.Unlock()

	return inner.RoundTrip(req)
}

// CloseIdleConnections closes idle connections owned by the underlying
// HTTP transport.
func (rt *utlsRoundTripper) CloseIdleConnections() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.close != nil {
		rt.close()
	}
}

func (rt *utlsRoundTripper) makeRoundTripper(
	req *http.Request,
) (http.RoundTripper, func(), error) {
	addr := req.URL.Host

	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "443")
	}

	// Bootstrap a connection so we can determine the negotiated ALPN.
	bootstrapConn, err := utlsDialContext(
		req.Context(),
		"tcp",
		addr,
		rt.targetIP,
		rt.config,
		rt.clientHelloID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap TLS connection: %w", err)
	}

	protocol := bootstrapConn.ConnectionState().NegotiatedProtocol

	dialTLSContext := rt.newDialTLSContext(
		addr,
		bootstrapConn,
		protocol,
	)

	switch rt.version {
	case HTTPVersionH1:
		tr := rt.newHTTP1Transport(dialTLSContext)

		return tr, tr.CloseIdleConnections, nil

	case HTTPVersionH2:
		tr := rt.newHTTP2Transport(dialTLSContext)

		return tr, tr.CloseIdleConnections, nil

	default:
		if protocol == http2.NextProtoTLS {
			tr := rt.newHTTP2Transport(dialTLSContext)

			return tr, tr.CloseIdleConnections, nil
		}

		tr := rt.newHTTP1Transport(dialTLSContext)

		return tr, tr.CloseIdleConnections, nil
	}
}

// newDialTLSContext creates the TLS dial function used by the underlying
// HTTP transport. The bootstrap connection is reused for the first dial.
func (rt *utlsRoundTripper) newDialTLSContext(
	addr string,
	bootstrapConn *utls.UConn,
	protocol string,
) func(context.Context, string, string) (net.Conn, error) {
	var mu sync.Mutex

	return func(ctx context.Context, network, _ string) (net.Conn, error) {
		mu.Lock()
		defer mu.Unlock()

		if bootstrapConn != nil {
			conn := bootstrapConn
			bootstrapConn = nil

			return conn, nil
		}

		conn, err := utlsDialContext(
			ctx,
			network,
			addr,
			rt.targetIP,
			rt.config,
			rt.clientHelloID,
		)
		if err != nil {
			return nil, err
		}

		negotiated := conn.ConnectionState().NegotiatedProtocol
		if negotiated != protocol {
			_ = conn.Close()

			return nil, fmt.Errorf(
				"unexpected switch from ALPN %q to %q",
				protocol,
				negotiated,
			)
		}

		return conn, nil
	}
}

func (rt *utlsRoundTripper) newHTTP1Transport(
	dialTLSContext func(context.Context, string, string) (net.Conn, error),
) *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()

	tr.DialTLSContext = dialTLSContext

	// Disable HTTP/2 upgrades.
	tr.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}

	return tr
}

func (rt *utlsRoundTripper) newHTTP2Transport(
	dialTLSContext func(context.Context, string, string) (net.Conn, error),
) *http2.Transport {
	return &http2.Transport{
		DialTLSContext: func(
			ctx context.Context,
			network, addr string,
			_ *tls.Config,
		) (net.Conn, error) {
			return dialTLSContext(ctx, network, addr)
		},
	}
}

// utlsHTTPClientFactory creates an HTTP client that uses uTLS for HTTPS
// connections to the specified target IP.
func utlsHTTPClientFactory(
	targetIP netip.Addr,
	timeout time.Duration,
	fingerprint string,
	minTLSVersion uint16,
	maxTLSVersion uint16,
	skipVerify bool,
	version HTTPVersion,
) (*http.Client, error) {
	clientHelloID := config.UTLSLookup(fingerprint)
	if clientHelloID == nil {
		return nil, fmt.Errorf("unknown uTLS fingerprint: %q", fingerprint)
	}

	utlsConfig := &utls.Config{
		InsecureSkipVerify: skipVerify,
		MinVersion:         minTLSVersion,
		MaxVersion:         maxTLSVersion,
	}

	rt := newUTLSRoundTripper(
		targetIP,
		utlsConfig,
		clientHelloID,
		version,
	)

	return &http.Client{
		Transport: rt,
		Timeout:   timeout,
	}, nil
}
