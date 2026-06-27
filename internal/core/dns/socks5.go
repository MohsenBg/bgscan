package dns

import (
	"context"
	"net"
	"net/http"
	"time"

	"bgscan/internal/logger"

	"golang.org/x/net/proxy"
)

// TestProxy performs a minimal HTTP connectivity check through a SOCKS5 proxy.
//
// A GET request is issued to the standard connectivity endpoint:
//
//	http://www.google.com/generate_204
//
// A successful proxy test is defined as receiving an HTTP 204 (No Content) status.
// The provided context controls cancellation and always takes priority. If a
// timeout duration is specified, it is applied as an upper bound on the total
// request time via context.WithTimeout.
//
// Returns true if the proxy is functional and the endpoint returns 204.
func TestProxy(ctx context.Context, proxyAddr string, timeout time.Duration) bool {
	// Reject immediately if the context is already cancelled.
	if err := ctx.Err(); err != nil {
		return false
	}

	// Apply timeout *on top of* existing ctx.
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Initialize SOCKS5 dialer.
	baseDialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		logger.CoreInfo("SOCKS5 dialer init failed for %s: %v", proxyAddr, err)
		return false
	}

	// Wrap the dialer with context support.
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return baseDialer.Dial(network, addr)
	}

	transport := &http.Transport{
		DialContext: dialContext,
	}

	client := &http.Client{
		Transport: transport,
		// Keep client-level timeout unset; ctx handles timing instead.
	}

	// Prepare request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://www.google.com/generate_204", nil)
	if err != nil {
		logger.CoreInfo("failed to create HTTP request via proxy %s: %v", proxyAddr, err)
		return false
	}

	// Perform request.
	resp, err := client.Do(req)
	if err != nil {
		logger.CoreInfo("proxy test request failed via %s: %v", proxyAddr, err)
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	// A valid connectivity check always returns HTTP 204.
	return resp.StatusCode == http.StatusNoContent
}
