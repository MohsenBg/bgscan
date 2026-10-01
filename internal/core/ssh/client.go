package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Client = ssh.Client

// handshakeTimeout bounds the SSH transport handshake so a hung peer
// can't block forever when the caller's context has no deadline.
const handshakeTimeout = 30 * time.Second

type SSHConfig struct {
	User           string
	Password       string
	PrivateKey     string
	KnownHostsFile string

	// HandshakeTimeout overrides the default handshake timeout.
	// Zero means use handshakeTimeout.
	HandshakeTimeout time.Duration

	// InsecureIgnoreHostKey disables host key verification when
	// KnownHostsFile is empty. Off by default to avoid silent MITM risk.
	InsecureIgnoreHostKey bool
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
//
// On any error (including context cancellation) the supplied conn is closed.
// On success the caller owns the returned *ssh.Client and is responsible for
// closing it.
func (s *sshService) Connect(
	ctx context.Context,
	conn net.Conn,
	addr string,
	config SSHConfig,
) (*Client, error) {
	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrContextCanceled, err)
	}

	auth, err := sshAuthMethods(config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	hostKeyCallback, err := hostKeyCallbackFor(config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         handshakeTimeoutFor(config),
	}

	type result struct {
		client *ssh.Client
		err    error
	}
	resultCh := make(chan result, 1)

	go func() {
		sshConn, chans, requests, err := ssh.NewClientConn(conn, addr, clientConfig)
		if err != nil {
			resultCh <- result{err: fmt.Errorf("%w: %w", ErrHandshake, err)}
			return
		}
		resultCh <- result{client: ssh.NewClient(sshConn, chans, requests)}
	}()

	select {
	case <-ctx.Done():
		_ = conn.Close()
		return nil, fmt.Errorf("%w: %w", ErrContextCanceled, ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			_ = conn.Close()
			return nil, res.err
		}
		return res.client, nil
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
			return nil, fmt.Errorf("%w: %w", ErrParsePrivateKey, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if len(methods) == 0 {
		return nil, ErrNoAuthConfigured
	}
	return methods, nil
}

func hostKeyCallbackFor(config SSHConfig) (ssh.HostKeyCallback, error) {
	if config.KnownHostsFile != "" {
		cb, err := knownhosts.New(config.KnownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrLoadKnownHosts, config.KnownHostsFile, err)
		}
		return cb, nil
	}
	if config.InsecureIgnoreHostKey {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	// No known_hosts and insecure mode not explicitly requested.
	return nil, fmt.Errorf("%w: known_hosts file is required", ErrLoadKnownHosts)
}

func handshakeTimeoutFor(config SSHConfig) time.Duration {
	if config.HandshakeTimeout > 0 {
		return config.HandshakeTimeout
	}
	return handshakeTimeout
}

// SSHDialContext returns a dial function suitable for use as a
// net.Dialer replacement that dials through the SSH client.
//
// On context cancellation the in-flight dial is left to complete in the
// background; the returned conn (if any) is closed to avoid a leak.
func (s *sshService) SSHDialContext(
	client *ssh.Client,
) func(context.Context, string, string) (net.Conn, error) {
	return func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrContextCanceled, err)
		}

		type result struct {
			conn net.Conn
			err  error
		}
		ch := make(chan result, 1)

		go func() {
			conn, err := client.Dial(network, address)
			if err != nil {
				ch <- result{err: fmt.Errorf("%w: %s %s: %w", ErrDial, network, address, err)}
				return
			}
			ch <- result{conn: conn}
		}()

		select {
		case <-ctx.Done():
			// Ensure the background dial's conn is closed if it succeeds.
			go func() {
				if res := <-ch; res.conn != nil {
					_ = res.conn.Close()
				}
			}()
			return nil, fmt.Errorf("%w: %w", ErrContextCanceled, ctx.Err())
		case res := <-ch:
			return res.conn, res.err
		}
	}
}

// DefaultKnownHostsPath returns the path to ~/.ssh/known_hosts when it
// exists, otherwise an empty string.
func DefaultKnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
