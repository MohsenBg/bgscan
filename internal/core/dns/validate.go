package dns

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	vaydns "github.com/net2share/vaydns/client"
	utls "github.com/refraction-networking/utls"
)

// parseClientHelloID resolves a case-insensitive TLS fingerprint label.
func parseClientHelloID(fingerprint string) (utls.ClientHelloID, error) {
	fingerprint = strings.TrimSpace(fingerprint)

	for _, client := range vaydns.UTLSClientHelloIDMap() {
		if strings.EqualFold(fingerprint, client.Label) {
			return *client.ID, nil
		}
	}

	return utls.ClientHelloID{}, fmt.Errorf(
		"unknown TLS fingerprint %q",
		fingerprint,
	)
}

// validateFingerprint validates a TLS fingerprint label.
func validateFingerprint(fingerprint string) error {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return fmt.Errorf("TLS fingerprint is required")
	}

	if _, err := parseClientHelloID(fingerprint); err != nil {
		return fmt.Errorf("invalid TLS fingerprint: %w", err)
	}

	return nil
}

// validateRPS validates a requests-per-second limit: 0 means unlimited.
func validateRPS(rps float64) error {
	if math.IsNaN(rps) || math.IsInf(rps, 0) || rps < 0 || rps > 500 {
		return fmt.Errorf("RPS must be between 0 and 500")
	}

	return nil
}

// validatePubKey validates a DNSTT/VayDNS public key:
// exactly 64 hexadecimal characters.
func validatePubKey(pubKey string) error {
	pubKey = strings.TrimSpace(pubKey)
	if pubKey == "" {
		return fmt.Errorf("public key is required")
	}

	if len(pubKey) != 64 {
		return fmt.Errorf("public key must be 64 hexadecimal characters")
	}

	if _, err := hex.DecodeString(pubKey); err != nil {
		return fmt.Errorf("public key must be hexadecimal")
	}

	return nil
}

// validatePrivateKey validates an SSH private key (PEM-encoded).
func validatePrivateKey(privateKey string) error {
	pemBlock := strings.TrimSpace(privateKey)
	if pemBlock == "" {
		return fmt.Errorf("private key is required")
	}

	if _, err := ssh.ParsePrivateKey([]byte(pemBlock)); err != nil {
		return fmt.Errorf("invalid SSH private key: %w", err)
	}

	return nil
}

func validateKnownHostsFile(path string) error {
	if path == "" {
		return fmt.Errorf("known hosts file is not configured")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("access known hosts file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("known hosts path is a directory")
	}

	if _, err := knownhosts.New(path); err != nil {
		return fmt.Errorf("invalid known hosts file: %w", err)
	}

	return nil
}

// validateProxyAuth validates the proxy and authentication fields shared
// by the tunnel configurations. An empty proxyType disables the proxy.
func validateProxyAuth(
	proxyType ResolverProxyType,
	proxyPort uint16,
	authMethod AuthMethod,
	username, password, privateKey, knownHostsFile string,
) map[string]error {
	errs := make(map[string]error)

	if proxyType != "" {
		if proxyPort == 0 {
			errs["proxy_port"] = fmt.Errorf("proxy port must be greater than zero")
		}

		switch proxyType {
		case ResolverProxySSH:
			if authMethod == AuthNone {
				errs["auth_method"] = fmt.Errorf("authentication is required for SSH proxy")
			}
		case ResolverProxySOCKS:
			if authMethod == AuthKey {
				errs["auth_method"] = fmt.Errorf("key auth is not supported for SOCKS proxy")
			}
		default:
			errs["proxy_type"] = fmt.Errorf("proxy type must be socks or ssh")
		}
	}

	if authMethod == AuthPassword {
		if strings.TrimSpace(username) == "" {
			errs["username"] = fmt.Errorf("username is required for password auth")
		}
		if strings.TrimSpace(password) == "" {
			errs["password"] = fmt.Errorf("password is required for password auth")
		}
	}

	if authMethod == AuthKey {
		if strings.TrimSpace(username) == "" {
			errs["username"] = fmt.Errorf("username is required for key auth")
		}
		if err := validatePrivateKey(privateKey); err != nil {
			errs["private_key"] = err
		}
		if knownHostsFile != "" {
			if err := validateKnownHostsFile(knownHostsFile); err != nil {
				errs["known_hosts_file"] = err
			}
		}
	}

	return errs
}
