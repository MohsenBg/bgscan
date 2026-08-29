package dns

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh"

	vaydns "github.com/net2share/vaydns/client"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/crypto/ssh/knownhosts"
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

// normalizeConfigName strips a .toml extension (case-insensitive) if present.
func normalizeConfigName(name string) string {
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".toml") {
		return strings.TrimSuffix(name, ext)
	}

	return name
}

func GetAllDNSTunsFile() ([]DNSTunConfigFile, error) {
	vaydnsCfg, err := NewVayDNSService().GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	dnsttCfg, err := NewDNSTTService().GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	slipstreamSrv, err := NewSlipstreamService()
	if err != nil {
		return nil, err
	}

	slipstreamCfg, err := slipstreamSrv.GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	configs := make(
		[]DNSTunConfigFile, 0,
		len(vaydnsCfg)+len(dnsttCfg)+len(slipstreamCfg),
	)

	for _, file := range vaydnsCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolVayDNS,
			Proxy:     string(file.Config.ProxyType) + "-" + string(file.Config.AuthMethod),
			Config:    file.Config,
		})
	}

	for _, file := range dnsttCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Proxy:     string(file.Config.ProxyType) + "-" + string(file.Config.AuthMethod),
			Protocol:  DNSTunProtocolDNSTT,
			Config:    file.Config,
		})
	}

	for _, file := range slipstreamCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolSlipstream,
			Proxy:     string(file.Config.ProxyType) + "-" + string(file.Config.AuthMethod),
			Config:    file.Config,
		})
	}

	slices.SortFunc(configs, func(a, b DNSTunConfigFile) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return configs, nil
}

func RenameDNSTunConfigFile(file DNSTunConfigFile, newName string) error {
	switch file.Protocol {
	case DNSTunProtocolVayDNS:
		return NewVayDNSService().RenameConfig(file.Name, newName)

	case DNSTunProtocolDNSTT:
		return NewDNSTTService().RenameConfig(file.Name, newName)

	case DNSTunProtocolSlipstream:
		srv, err := NewSlipstreamService()
		if err != nil {
			return err
		}

		return srv.RenameConfig(file.Name, newName)

	default:
		return fmt.Errorf("unsupported DNS tunnel protocol: %q", file.Protocol)
	}
}
