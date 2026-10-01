package ssh

import "errors"

var (
	// ErrContextCanceled is returned when the provided context is canceled
	// before the SSH handshake or dial completes.
	ErrContextCanceled = errors.New("ssh: context canceled")

	// ErrNoAuthConfigured is returned when neither Password nor PrivateKey
	// was supplied in SSHConfig.
	ErrNoAuthConfigured = errors.New("ssh: authentication is not configured")

	// ErrParsePrivateKey is returned when the supplied private key cannot
	// be parsed into a valid ssh.Signer.
	ErrParsePrivateKey = errors.New("ssh: parse private key")

	// ErrLoadKnownHosts is returned when the known_hosts file cannot be
	// loaded or parsed.
	ErrLoadKnownHosts = errors.New("ssh: load known_hosts")

	// ErrHandshake is returned when the SSH transport handshake fails
	// (version exchange, kex, host key verification, or auth).
	ErrHandshake = errors.New("ssh: handshake")

	// ErrDial is returned when dialing through an established SSH client
	// fails (channel open / forwarding rejected).
	ErrDial = errors.New("ssh: dial")
)
