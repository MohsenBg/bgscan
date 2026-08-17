package ssh

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestSSHService_Connect_Success(t *testing.T) {
	serverConfig, err := newTestServerConfig("testuser", "secret123")
	if err != nil {
		t.Fatalf("newTestServerConfig() error = %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	serverErr := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		serverErr <- runMockServer(conn, serverConfig)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial() error = %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	service := NewSSHService()

	client, err := service.Connect(
		t.Context(),
		conn,
		listener.Addr().String(),
		SSHConfig{
			User:     "testuser",
			Password: "secret123",
		},
	)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if client == nil {
		t.Fatal("Connect() returned nil client")
	}

	if err := client.Close(); err != nil {
		t.Errorf("client.Close() error = %v", err)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("mock server error = %v", err)
	}
}

func TestSSHService_Connect_AuthenticationFailure(t *testing.T) {
	serverConfig, err := newTestServerConfig("testuser", "secret123")
	if err != nil {
		t.Fatalf("newTestServerConfig() error = %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	serverErr := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		serverErr <- runMockServer(conn, serverConfig)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial() error = %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	service := NewSSHService()

	client, err := service.Connect(
		t.Context(),
		conn,
		listener.Addr().String(),
		SSHConfig{
			User:     "wronguser",
			Password: "wrongpassword",
		},
	)
	if err == nil {
		if client != nil {
			_ = client.Close()
		}

		t.Fatal("Connect() expected authentication error")
	}

	if client != nil {
		_ = client.Close()
		t.Fatal("Connect() returned client on authentication failure")
	}

	if serverErr := <-serverErr; serverErr == nil {
		t.Fatal("mock server expected authentication error")
	}
}

// runMockServer performs the SSH server handshake and keeps the
// connection alive long enough for the client to complete the test.
func runMockServer(conn net.Conn, config *ssh.ServerConfig) error {
	serverConn, chans, requests, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return err
	}
	defer func() {
		_ = serverConn.Close()
	}()

	go ssh.DiscardRequests(requests)

	for newChannel := range chans {
		_ = newChannel.Reject(
			ssh.UnknownChannelType,
			"unknown channel type",
		)
	}

	return nil
}

func newTestServerConfig(
	user string,
	password string,
) (*ssh.ServerConfig, error) {
	privateKey, err := ecdsa.GenerateKey(
		elliptic.P256(),
		rand.Reader,
	)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, err
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(
			conn ssh.ConnMetadata,
			pass []byte,
		) (*ssh.Permissions, error) {
			if conn.User() != user || string(pass) != password {
				return nil, ssh.ErrNoAuth
			}

			return nil, nil
		},
	}

	config.AddHostKey(signer)

	return config, nil
}
