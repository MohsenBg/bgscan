package speedtest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type httpClientConfig struct {
	ProxyPort   uint16
	DialContext func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error)
}

const connectTimeout = 10 * time.Second

func newHTTPClient(cfg httpClientConfig) (*http.Client, error) {
	transport := &http.Transport{
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   connectTimeout,
		ResponseHeaderTimeout: connectTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	switch {
	case cfg.DialContext != nil:
		transport.DialContext = func(
			ctx context.Context,
			network string,
			address string,
		) (net.Conn, error) {
			dialCtx, cancel := context.WithTimeout(ctx, connectTimeout)
			defer cancel()
			return cfg.DialContext(dialCtx, network, address)
		}

	case cfg.ProxyPort == 0:
		return nil, fmt.Errorf(
			"speedtest: DialContext or ProxyPort must be set",
		)

	default:
		proxyURL, err := url.Parse(
			fmt.Sprintf("socks5://127.0.0.1:%d", cfg.ProxyPort),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"speedtest: invalid proxy URL on port %d: %w",
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
