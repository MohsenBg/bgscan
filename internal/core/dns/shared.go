package dns

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	vaydns "github.com/net2share/vaydns/client"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/idna"
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

// validateDomain validates a domain name: non-empty, valid IDNA conversion,
// no leading/trailing dots, no empty labels, valid label syntax and length.
func validateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}

	ascii, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return fmt.Errorf("domain is invalid: %w", err)
	}

	if strings.HasPrefix(ascii, ".") ||
		strings.HasSuffix(ascii, ".") ||
		strings.Contains(ascii, "..") {
		return fmt.Errorf("domain is invalid")
	}

	labels := strings.Split(ascii, ".")
	if len(labels) < 2 {
		return fmt.Errorf("domain must contain at least two labels")
	}

	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return fmt.Errorf("domain is invalid")
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("domain is invalid")
		}
		for _, r := range label {
			if (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-' {
				continue
			}
			return fmt.Errorf("domain is invalid")
		}
	}

	if len(ascii) > 253 {
		return fmt.Errorf("domain is invalid")
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
