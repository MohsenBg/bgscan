package httpprobe

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"testing"
	"time"
)

// startRecordingServer serves requests on ln and reports the Host header of
// the first request it receives.
func startRecordingServer(t *testing.T, ln net.Listener, hit chan<- string) {
	t.Helper()

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case hit <- r.Host:
		default:
		}
		w.WriteHeader(http.StatusOK)
	})}

	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
}

func TestUTLSHTTPClient_PlainHTTPConnectsToTargetIP(t *testing.T) {
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on target address: %v", err)
	}

	port := targetLn.Addr().(*net.TCPAddr).Port

	unrelatedLn, err := net.Listen("tcp", net.JoinHostPort("127.0.0.2", strconv.Itoa(port)))
	if err != nil {
		_ = targetLn.Close()
		t.Skipf("cannot bind unrelated loopback address: %v", err)
	}

	targetHit := make(chan string, 1)
	unrelatedHit := make(chan string, 1)
	startRecordingServer(t, targetLn, targetHit)
	startRecordingServer(t, unrelatedLn, unrelatedHit)

	client, err := utlsHTTPClientFactory(
		netip.MustParseAddr("127.0.0.1"),
		2*time.Second,
		"Chrome",
		0,
		0,
		true,
		HTTPVersionH1,
	)
	if err != nil {
		t.Fatalf("create uTLS client: %v", err)
	}
	defer client.CloseIdleConnections()

	url := fmt.Sprintf("http://127.0.0.2:%d/plain", port)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	select {
	case host := <-targetHit:
		wantHost := fmt.Sprintf("127.0.0.2:%d", port)
		if host != wantHost {
			t.Fatalf("target received Host %q, want %q", host, wantHost)
		}
	default:
		t.Fatal("target IP listener did not receive the request")
	}

	select {
	case host := <-unrelatedHit:
		t.Fatalf("request reached the unrelated host %q", host)
	default:
	}
}
