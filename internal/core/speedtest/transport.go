package speedtest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// connectTimeout is a fixed budget for TCP dial + TLS handshake.
// Kept separate from the transfer timeout so connection latency
// does not eat into the speed measurement window.
const connectTimeout = 10 * time.Second

func newHTTPClient(port uint16) (*http.Client, error) {
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://127.0.0.1:%d", port))
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL on port %d: %w", port, err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: connectTimeout,
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
	}

	return &http.Client{Transport: transport}, nil
}

// contextReader wraps an io.Reader and checks context cancellation on
// every Read so io.Copy respects the transfer deadline.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}
