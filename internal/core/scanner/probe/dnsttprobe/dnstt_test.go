package dnsttprobe

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/process"
	"bgscan/internal/core/scanner/probe"
)

type fakePortManager struct {
	port     uint16
	getErr   error
	released []uint16
}

func (m *fakePortManager) Get(context.Context) (uint16, error) {
	return m.port, m.getErr
}

func (m *fakePortManager) Release(port uint16) {
	m.released = append(m.released, port)
}

func (m *fakePortManager) Close() {}

type fakeProcess struct {
	stopCalled bool
}

func (f *fakeProcess) StopGracefully(time.Duration) error { f.stopCalled = true; return nil }
func (f *fakeProcess) Kill() error                        { f.stopCalled = true; return nil }
func (*fakeProcess) Wait() error                          { return nil }

type fakeProcessTracker struct {
	started         bool
	registerErr     error
	unregisterErr   error
	registeredID    string
	unregisteredIDs []string
}

func (t *fakeProcessTracker) Start(context.Context) {
	t.started = true
}

func (t *fakeProcessTracker) Register(context.Context, probe.Killable) (string, error) {
	return t.registeredID, t.registerErr
}

func (t *fakeProcessTracker) Unregister(_ context.Context, id string) error {
	t.unregisteredIDs = append(t.unregisteredIDs, id)
	return t.unregisterErr
}

type fakeDNSTTClient struct {
	process process.Process
	runErr  error

	runCalled bool
	runIP     string
	runPort   uint16
}

func (c *fakeDNSTTClient) RunTunnel(_ context.Context, ip string, port uint16) (process.Process, error) {
	c.runCalled = true
	c.runIP = ip
	c.runPort = port

	return c.process, c.runErr
}

func validConfig() DNSTTConfig {
	return DNSTTConfig{
		Domain:    "tunnel.example.com",
		PubKey:    "test-public-key",
		Transport: dns.UDP,
		DNSPort:   53,
		Timeout:   2 * time.Second,
	}
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(t *testing.T, config DNSTTConfig, pm *fakePortManager, tracker *fakeProcessTracker, client *fakeDNSTTClient) *DNSTTProbe {
	t.Helper()

	got, err := NewDNSTTProbe(
		config,
		pm,
		WithProcessTracker(tracker),
		WithClient(client),
	)
	if err != nil {
		t.Fatal(err)
	}

	p, ok := got.(*DNSTTProbe)
	if !ok {
		t.Fatalf("NewDNSTTProbe returned %T, want *DNSTTProbe", got)
	}

	p.waitOpen = func(context.Context, string, time.Duration) error { return nil }
	p.testProxy = func(context.Context, string, time.Duration) bool { return true }

	return p
}

func TestNewDNSTTProbeValidation(t *testing.T) {
	client := &fakeDNSTTClient{}

	if _, err := NewDNSTTProbe(validConfig(), nil, WithClient(client)); err == nil {
		t.Fatal("expected an error for a nil port manager")
	}

	config := validConfig()
	config.Domain = ""

	if _, err := NewDNSTTProbe(config, &fakePortManager{}, WithClient(client)); err == nil {
		t.Fatal("expected an error for an empty domain")
	}
}

func TestNewDNSTTProbeAppliesDefaultTimeoutAndOptions(t *testing.T) {
	config := validConfig()
	config.Timeout = 0

	pm := &fakePortManager{}
	tracker := &fakeProcessTracker{}
	client := &fakeDNSTTClient{}

	got, err := NewDNSTTProbe(
		config,
		pm,
		WithProcessTracker(tracker),
		WithClient(client),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*DNSTTProbe)

	if p.config.Timeout != 5*time.Second {
		t.Fatalf("timeout = %s, want 5s", p.config.Timeout)
	}

	if p.processTracker != tracker {
		t.Fatal("process tracker option was not applied")
	}

	if p.client != client {
		t.Fatal("client option was not applied")
	}
}

func TestInitStartsProcessTracker(t *testing.T) {
	tracker := &fakeProcessTracker{}
	p := newTestProbe(t, validConfig(), &fakePortManager{}, tracker, &fakeDNSTTClient{})

	if err := p.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !tracker.started {
		t.Fatal("process tracker was not started")
	}
}

func TestRunReturnsCanceledContextBeforeAllocatingPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pm := &fakePortManager{port: 4000}
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{}, &fakeDNSTTClient{})

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenAllocationFails(t *testing.T) {
	pm := &fakePortManager{getErr: errors.New("no ports available")}
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{}, &fakeDNSTTClient{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an allocation error")
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenTunnelFails(t *testing.T) {
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 5000}
	client := &fakeDNSTTClient{process: prc, runErr: errors.New("start failed")}
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{}, client)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a tunnel error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 5000 {
		t.Fatalf("released ports = %v, want [5000]", got)
	}
}

func TestRunStopsTunnelWhenRegistrationFails(t *testing.T) {
	pm := &fakePortManager{port: 6000}
	tracker := &fakeProcessTracker{registerErr: errors.New("register failed")}
	client := &fakeDNSTTClient{process: &fakeProcess{}}
	p := newTestProbe(t, validConfig(), pm, tracker, client)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a registration error")
	}

	if len(tracker.unregisteredIDs) != 0 {
		t.Fatalf("unregistered IDs = %v, want none", tracker.unregisteredIDs)
	}
}

func TestRunCleansUpAfterWaitOpenFailure(t *testing.T) {
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 7000}
	tracker := &fakeProcessTracker{registeredID: "process-1"}
	client := &fakeDNSTTClient{process: prc}
	p := newTestProbe(t, validConfig(), pm, tracker, client)
	p.waitOpen = func(context.Context, string, time.Duration) error {
		return errors.New("proxy did not open")
	}

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a wait-open error")
	}

	if !prc.stopCalled {
		t.Fatal("StopTunnel was not called")
	}

	if got := tracker.unregisteredIDs; len(got) != 1 || got[0] != "process-1" {
		t.Fatalf("unregistered IDs = %v, want [process-1]", got)
	}

	if got := pm.released; len(got) != 1 || got[0] != 7000 {
		t.Fatalf("released ports = %v, want [7000]", got)
	}
}

func TestRunReturnsContextErrorWhenProxyCheckCancelsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracker := &fakeProcessTracker{registeredID: "process-1"}
	client := &fakeDNSTTClient{process: &fakeProcess{}}
	p := newTestProbe(t, validConfig(), &fakePortManager{port: 8000}, tracker, client)
	p.testProxy = func(context.Context, string, time.Duration) bool {
		cancel()
		return false
	}

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRunSuccess(t *testing.T) {
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 9000}
	tracker := &fakeProcessTracker{registeredID: "process-1"}
	client := &fakeDNSTTClient{process: prc}
	p := newTestProbe(t, validConfig(), pm, tracker, client)

	var waitOpenAddr string
	p.waitOpen = func(_ context.Context, addr string, timeout time.Duration) error {
		waitOpenAddr = addr

		if timeout != time.Second {
			t.Errorf("wait timeout = %s, want 1s", timeout)
		}

		return nil
	}

	var proxyAddr string
	var proxyTimeout time.Duration
	p.testProxy = func(_ context.Context, addr string, timeout time.Duration) bool {
		proxyAddr = addr
		proxyTimeout = timeout
		return true
	}

	result, err := p.Run(context.Background(), testIP())
	if err != nil {
		t.Fatal(err)
	}

	got, ok := result.(DNSTTResult)
	if !ok {
		t.Fatalf("result type = %T, want DNSTTResult", result)
	}

	if got.IP != testIP() || got.Port != 9000 || got.Transport != dns.UDP {
		t.Fatalf("unexpected result: %#v", got)
	}

	if got.Latency < 0 {
		t.Fatalf("latency = %s, want a non-negative value", got.Latency)
	}

	if !client.runCalled || client.runIP != "1.2.3.4" || client.runPort != 9000 {
		t.Fatalf("RunTunnel call = (%q, %d), want (1.2.3.4, 9000)", client.runIP, client.runPort)
	}

	if waitOpenAddr != "127.0.0.1:9000" || proxyAddr != "127.0.0.1:9000" {
		t.Fatalf("proxy addresses = (%q, %q), want 127.0.0.1:9000", waitOpenAddr, proxyAddr)
	}

	if proxyTimeout != 2*time.Second {
		t.Fatalf("proxy timeout = %s, want 2s", proxyTimeout)
	}

	if !prc.stopCalled {
		t.Fatal("StopTunnel was not called")
	}

	if got := tracker.unregisteredIDs; len(got) != 1 || got[0] != "process-1" {
		t.Fatalf("unregistered IDs = %v, want [process-1]", got)
	}

	if got := pm.released; len(got) != 1 || got[0] != 9000 {
		t.Fatalf("released ports = %v, want [9000]", got)
	}
}

func TestClose(t *testing.T) {
	if err := (&DNSTTProbe{}).Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
