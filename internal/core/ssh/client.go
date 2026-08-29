package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client = ssh.Client

type SSHConfig struct {
	User           string
	Password       string
	PrivateKey     string
	KnownHostsFile string
}

type SSHService interface {
	Connect(ctx context.Context, conn net.Conn, addr string, config SSHConfig) (*ssh.Client, error)
	SSHDialContext(client *ssh.Client) func(context.Context, string, string) (net.Conn, error)
}

type sshService struct{}

func NewSSHService() SSHService {
	return &sshService{}
}

// Connect performs the SSH handshake and authentication over conn.

func (s *sshService) Connect(
	ctx context.Context,
	conn net.Conn,
	addr string,
	config SSHConfig,
) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	auth, err := sshAuthMethods(config)
	if err != nil {
		return nil, err
	}

	var hostKeyCallback ssh.HostKeyCallback
	if config.KnownHostsFile != "" {
		var err error
		hostKeyCallback, err = knownhosts.New(config.KnownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load known_hosts: %w", err)
		}
	} else {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
	}

	type result struct {
		client *ssh.Client
		err    error
	}
	resultCh := make(chan result, 1)
	go func() {
		sshConn, chans, requests, err := ssh.NewClientConn(conn, addr, clientConfig)
		if err != nil {
			resultCh <- result{err: fmt.Errorf("SSH handshake: %w", err)}
			return
		}
		resultCh <- result{client: ssh.NewClient(sshConn, chans, requests)}
	}()

	select {
	case <-ctx.Done():
		_ = conn.Close()
		return nil, ctx.Err()
	case res := <-resultCh:
		return res.client, res.err
	}
}

func sshAuthMethods(config SSHConfig) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if config.Password != "" {
		methods = append(methods, ssh.Password(config.Password))
	}
	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("parse SSH private key: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("SSH authentication is not configured")
	}
	return methods, nil
}

func (s *sshService) SSHDialContext(
	client *ssh.Client,
) func(context.Context, string, string) (net.Conn, error) {
	return func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		type result struct {
			conn net.Conn
			err  error
		}
		ch := make(chan result, 1)
		go func() {
			conn, err := client.Dial(network, address)
			ch <- result{conn: conn, err: err}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-ch:
			return result.conn, result.err
		}
	}
}

func DefaultKnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		path := filepath.Join(home, ".ssh", "known_hosts")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return home
}
