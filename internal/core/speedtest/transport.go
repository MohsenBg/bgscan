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

// connectTimeout limits TCP dialing and TLS handshakes.
const connectTimeout = 10 * time.Second

type httpClientConfig struct {
	ProxyPort   uint16
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}

func newHTTPClient(cfg httpClientConfig) (*http.Client, error) {
	transport := &http.Transport{
		TLSHandshakeTimeout: connectTimeout,
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
	}

	if cfg.DialContext != nil {
		transport.DialContext = cfg.DialContext
	} else {
		proxyURL, err := url.Parse(
			fmt.Sprintf("socks5://127.0.0.1:%d", cfg.ProxyPort),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid proxy URL on port %d: %w",
				cfg.ProxyPort,
				err,
			)
		}

		transport.Proxy = http.ProxyURL(proxyURL)
		transport.DialContext = (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	return &http.Client{Transport: transport}, nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	return r.r.Read(p)
}
