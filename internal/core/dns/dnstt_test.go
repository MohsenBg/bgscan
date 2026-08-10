package dns

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"bgscan/internal/core/process"
)

type mockProcess struct {
	stopErr error
	killErr error
	waitErr error

	stopCalled bool
	killCalled bool
	waitCalled bool

	stopTimeout time.Duration
}

func (p *mockProcess) StopGracefully(timeout time.Duration) error {
	p.stopCalled = true
	p.stopTimeout = timeout

	return p.stopErr
}

func (p *mockProcess) Kill() error {
	p.killCalled = true

	return p.killErr
}

func (p *mockProcess) Wait() error {
	p.waitCalled = true

	return p.waitErr
}

func TestNewDNSTTClient(t *testing.T) {
	t.Parallel()

	client, err := NewDNSTTClient(
		"example.com",
		"public-key",
		UDP,
		53,
		WithDNSTTClientBinary("test-dnstt-client"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if client == nil {
		t.Fatal("NewDNSTTClient returned nil client")
	}
}

func TestRunTunnel(t *testing.T) {
	t.Parallel()

	wantProc := &mockProcess{}

	var gotBin string
	var gotArgs []string
	var startCalls int

	client, err := NewDNSTTClient(
		"example.com",
		"public-key",
		UDP,
		53,
		WithDNSTTClientBinary("test-dnstt-client"),
		WithProcessStarter(func(
			_ context.Context,
			bin string,
			args ...string,
		) (process.Process, error) {
			startCalls++
			gotBin = bin
			gotArgs = append([]string(nil), args...)

			return wantProc, nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	gotProc, err := client.RunTunnel(context.Background(), "8.8.8.8", 1080)
	if err != nil {
		t.Fatal(err)
	}

	if gotProc != wantProc {
		t.Fatalf("RunTunnel() process = %v, want %v", gotProc, wantProc)
	}

	if startCalls != 1 {
		t.Fatalf("starter called %d times, want 1", startCalls)
	}

	if gotBin != "test-dnstt-client" {
		t.Fatalf("binary = %q, want %q", gotBin, "test-dnstt-client")
	}

	wantArgs := []string{
		"-udp",
		"8.8.8.8:53",
		"-pubkey",
		"public-key",
		"example.com",
		"127.0.0.1:1080",
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("arguments = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestRunTunnelReturnsStarterError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("start failed")

	client, err := NewDNSTTClient(
		"example.com",
		"public-key",
		UDP,
		53,
		WithDNSTTClientBinary("test-dnstt-client"),
		WithProcessStarter(func(
			context.Context,
			string,
			...string,
		) (process.Process, error) {
			return nil, wantErr
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	proc, err := client.RunTunnel(context.Background(), "8.8.8.8", 1080)

	if proc != nil {
		t.Fatalf("process = %v, want nil", proc)
	}

	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestGetDNSTransportFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		transport Transport
		want      string
	}{
		{
			name:      "UDP",
			transport: UDP,
			want:      "-udp",
		},
		{
			name:      "TCP",
			transport: TCP,
			want:      "-dot",
		},
		{
			name:      "DOT",
			transport: DOT,
			want:      "-dot",
		},
		{
			name:      "DOH falls back to DOT",
			transport: DOH,
			want:      "-dot",
		},
		{
			name:      "unknown falls back to UDP",
			transport: Transport("unknown"),
			want:      "-udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDNSTransportFlag(tt.transport); got != tt.want {
				t.Fatalf("getDNSTransportFlag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDNSTTClientPaths(t *testing.T) {
	t.Parallel()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}

	base := filepath.Dir(exe)

	want := []string{
		filepath.Join(base, "assets", "dnstt-client"),
		filepath.Join(base, "assets", "dns", "dnstt-client"),
		filepath.Join(base, "dnstt-client"),
		base,
	}

	if got := getDNSTTPaths(); !reflect.DeepEqual(got, want) {
		t.Fatalf("getDNSTTPaths() = %#v, want %#v", got, want)
	}
}
