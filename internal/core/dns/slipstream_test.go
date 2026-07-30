package dns

import (
	"context"
	"reflect"
	"testing"

	"bgscan/internal/core/process"
)

func TestSlipstreamRunTunnel(t *testing.T) {
	wantProc := &mockProcess{}

	var gotBin string
	var gotArgs []string

	client, err := NewSlipstreamClient(
		"example.com",
		53,
		"cert.pem",
		WithSlipstreamClientBinary("test-slipstream-client"),
		WithSlipstreamProcessStarter(func(
			_ context.Context,
			bin string,
			args ...string,
		) (process.Process, error) {
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
		t.Fatalf("process = %v, want %v", gotProc, wantProc)
	}

	if gotBin != "test-slipstream-client" {
		t.Fatalf("binary = %q", gotBin)
	}

	wantArgs := []string{
		"-d", "example.com",
		"-r", "8.8.8.8:53",
		"-l", "1080",
		"--cert", "cert.pem",
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}
