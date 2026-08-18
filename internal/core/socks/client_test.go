package socks

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeSOCKSServer implements the minimal server side of the SOCKS5
// protocol, recording what the client sends.
type fakeSOCKSServer struct {
	greeting      []byte
	authRequest   []byte
	connectReq    []byte
	methodReply   byte
	authStatus    byte
	connectStatus byte

	// When set, Connect replies with these bytes instead of a normal
	// method-selection response.
	rawMethodReply []byte
}

func (s *fakeSOCKSServer) Serve(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	buf := make([]byte, 3)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}
	s.greeting = append([]byte(nil), buf...)

	if s.rawMethodReply != nil {
		if _, err := conn.Write(s.rawMethodReply); err != nil {
			return
		}
	} else {
		s.methodReply = s.greeting[2]
		if _, err := conn.Write([]byte{0x05, s.greeting[2]}); err != nil {
			return
		}
	}

	if s.greeting[2] == 0x02 {
		// Read username/password auth: VER ULEN UNAME PLEN PASSWD.
		header := make([]byte, 2)
		if _, err := io.ReadFull(conn, header); err != nil {
			return
		}
		user := make([]byte, header[1])
		if _, err := io.ReadFull(conn, user); err != nil {
			return
		}
		passLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, passLen); err != nil {
			return
		}
		pass := make([]byte, passLen[0])
		if _, err := io.ReadFull(conn, pass); err != nil {
			return
		}
		s.authRequest = append([]byte{header[0], header[1]}, user...)
		s.authRequest = append(s.authRequest, passLen[0])
		s.authRequest = append(s.authRequest, pass...)

		s.authStatus = 0x00
		if string(pass) == "wrong" {
			s.authStatus = 0x01
		}
		if _, err := conn.Write([]byte{0x01, s.authStatus}); err != nil {
			return
		}
		if s.authStatus != 0x00 {
			return
		}
	}

	req := make([]byte, 4)
	if _, err := io.ReadFull(conn, req); err != nil {
		return
	}
	switch req[3] {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		req = append(req, addr...)
	case 0x03: // Domain
		l := make([]byte, 1)
		if _, err := io.ReadFull(conn, l); err != nil {
			return
		}
		addr := make([]byte, l[0])
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		req = append(req, l[0])
		req = append(req, addr...)
	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		req = append(req, addr...)
	default:
		return
	}
	port := make([]byte, 2)
	if _, err := io.ReadFull(conn, port); err != nil {
		return
	}
	req = append(req, port...)

	s.connectReq = req

	// Reply: VER REP RSV ATYP ADDR.PORT.
	reply := []byte{0x05, s.connectStatus, 0x00, 0x01, 127, 0, 0, 1, 0x00, 0x50}
	if _, err := conn.Write(reply); err != nil {
		return
	}

	// Keep the connection open until the client is done.
	for {
		if _, err := conn.Read(buf); err != nil {
			return
		}
	}
}

// serveFakeSOCKS starts a fake SOCKS5 server on a random local port and
// returns its address.
func serveFakeSOCKS(t *testing.T, server *fakeSOCKSServer) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		server.Serve(conn)
	}()

	return ln.Addr().String()
}

func dialTest(t *testing.T, addr string) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial %q: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func TestConnectNoAuth(t *testing.T) {
	server := &fakeSOCKSServer{}
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn := dialTest(t, addr)
	got, err := svc.Connect(ctx, conn, "93.184.216.34:443", Config{})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	_ = got.Close()

	if server.greeting[0] != 0x05 {
		t.Errorf("greeting version = %d, want 5", server.greeting[0])
	}
	if server.methodReply != 0x00 {
		t.Errorf("selected method = 0x%02x, want 0x00 (no auth)", server.methodReply)
	}

	// Request: VER=05 CMD=01 RSV=00 ATYP=01 IPv4.
	if len(server.connectReq) < 10 {
		t.Fatalf("connect request too short: %d bytes", len(server.connectReq))
	}
	if server.connectReq[0] != 0x05 || server.connectReq[1] != 0x01 || server.connectReq[2] != 0x00 {
		t.Errorf("request header = % x, want VER=05 CMD=01 RSV=00", server.connectReq[:3])
	}
	if server.connectReq[3] != 0x01 {
		t.Errorf("address type = %#02x, want 0x01 (IPv4)", server.connectReq[3])
	}
	ip := net.IP(server.connectReq[4:8]).String()
	if ip != "93.184.216.34" {
		t.Errorf("encoded address = %q, want 93.184.216.34", ip)
	}
	port := binary.BigEndian.Uint16(server.connectReq[len(server.connectReq)-2:])
	if port != 443 {
		t.Errorf("port = %d, want 443", port)
	}
}

func TestConnectUsernamePasswordAuth(t *testing.T) {
	server := &fakeSOCKSServer{}
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn := dialTest(t, addr)
	if _, err := svc.Connect(ctx, conn, "example.com:80", Config{User: "alice", Password: "secret"}); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if server.methodReply != 0x02 {
		t.Errorf("selected method = 0x%02x, want 0x02 (username/password)", server.methodReply)
	}
	if server.authStatus != 0x00 {
		t.Errorf("auth status = %#02x, want 0x00", server.authStatus)
	}

	// The auth sub-negotiation must carry our credentials verbatim.
	got := string(server.authRequest)
	if !strings.Contains(got, "alice") || !strings.Contains(got, "secret") {
		t.Errorf("auth payload = %q, want it to contain credentials", got)
	}

	// Target uses ATYP 0x03 for hostnames.
	if server.connectReq[3] != 0x03 {
		t.Errorf("address type = %#02x, want 0x03 (domain)", server.connectReq[3])
	}
	domainLen := int(server.connectReq[4])
	if string(server.connectReq[5:5+domainLen]) != "example.com" {
		t.Errorf(
			"domain = %q, want example.com",
			server.connectReq[5:5+domainLen],
		)
	}
}

func TestConnectAuthRejected(t *testing.T) {
	server := &fakeSOCKSServer{
		// Server accepts no auth method; client offers username/password.
		rawMethodReply: []byte{0x05, 0xff},
	}
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	conn := dialTest(t, addr)

	_, err := svc.Connect(context.Background(), conn, "example.com:80", Config{User: "u", Password: "p"})
	if err == nil {
		t.Fatal("expected error when no acceptable auth method")
	}
	if !strings.Contains(err.Error(), "no acceptable authentication method") &&
		!strings.Contains(err.Error(), "authentication") {
		t.Errorf("error = %v, want an auth-related failure", err)
	}
}

func TestConnectWrongCredentialsRejected(t *testing.T) {
	server := &fakeSOCKSServer{}
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	conn := dialTest(t, addr)

	_, err := svc.Connect(context.Background(), conn, "example.com:80", Config{User: "alice", Password: "wrong"})
	if err == nil {
		t.Fatal("expected error for rejected credentials")
	}
	if !strings.Contains(err.Error(), "authentication rejected") {
		t.Errorf("error = %v, want 'authentication rejected'", err)
	}
}

func TestConnectServerRejectsCommand(t *testing.T) {
	server := &fakeSOCKSServer{}
	// 0x07 = command not supported in the CONNECT reply.
	server.connectStatus = 0x07
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	conn := dialTest(t, addr)

	_, err := svc.Connect(context.Background(), conn, "example.com:80", Config{})
	if err == nil {
		t.Fatal("expected error for failed CONNECT")
	}
	if !strings.Contains(err.Error(), "command not supported") {
		t.Errorf("error = %v, want 'command not supported'", err)
	}
}

func TestConnectInvalidTargetPort(t *testing.T) {
	server := &fakeSOCKSServer{}
	addr := serveFakeSOCKS(t, server)

	svc := NewService()
	conn := dialTest(t, addr)

	_, err := svc.Connect(context.Background(), conn, "example.com:notaport", Config{})
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
	if !errors.Is(err, err) || !strings.Contains(err.Error(), "invalid target") {
		t.Errorf("error = %v, want invalid-target error", err)
	}
}

func TestEncodeAddressIPv6(t *testing.T) {
	atyp, data, err := encodeAddress("2001:db8::1")
	if err != nil {
		t.Fatalf("encodeAddress() error = %v", err)
	}
	if atyp != 0x04 {
		t.Errorf("atyp = %#02x, want 0x04 (IPv6)", atyp)
	}
	ip := net.IP(data).String()
	if ip != "2001:db8::1" {
		t.Errorf("decoded address = %q, want 2001:db8::1", ip)
	}
}

func TestEncodeAddressDomainTooLong(t *testing.T) {
	long := strings.Repeat("a", 256)
	if _, _, err := encodeAddress(long); err == nil {
		t.Fatal("expected error for over-long hostname")
	}
}
