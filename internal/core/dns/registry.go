package dns

import (
	"fmt"
	"slices"
)

// GetAllDNSTunsFile returns all tunnel configuration files across every
// supported protocol, newest first.
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

	masterdnsCfg, err := NewMasterDNSService().GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	stormdnsCfg, err := NewStormDNSService().GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	thefeedCfg, err := NewTheFeedService().GetAllConfigFiles()
	if err != nil {
		return nil, err
	}

	configs := make(
		[]DNSTunConfigFile, 0,
		len(vaydnsCfg)+len(dnsttCfg)+len(slipstreamCfg)+len(masterdnsCfg)+len(stormdnsCfg)+len(thefeedCfg),
	)

	for _, file := range vaydnsCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolVayDNS,
			Proxy:     proxyLabel(file.Config.ProxyType, file.Config.AuthMethod),
			Config:    file.Config,
		})
	}

	for _, file := range dnsttCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolDNSTT,
			Proxy:     proxyLabel(file.Config.ProxyType, file.Config.AuthMethod),
			Config:    file.Config,
		})
	}

	for _, file := range slipstreamCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolSlipstream,
			Proxy:     proxyLabel(file.Config.ProxyType, file.Config.AuthMethod),
			Config:    file.Config,
		})
	}

	for _, file := range masterdnsCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolMasterDNS,
			Proxy:     "socks-" + file.Config.DataEncMethod.String(),
			Config:    file.Config,
		})
	}

	for _, file := range stormdnsCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolStormDNS,
			Proxy:     "socks-" + file.Config.DataEncMethod.String(),
			Config:    file.Config,
		})
	}

	for _, file := range thefeedCfg {
		configs = append(configs, DNSTunConfigFile{
			Name:      file.Name,
			Path:      file.Path,
			CreatedAt: file.CreatedAt,
			Protocol:  DNSTunProtocolTheFeed,
			Proxy:     "socks-" + file.Config.QueryMode,
			Config:    file.Config,
		})
	}

	slices.SortFunc(configs, func(a, b DNSTunConfigFile) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return configs, nil
}

// RenameDNSTunConfigFile renames a tunnel config file, dispatching on
// its protocol.
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

	case DNSTunProtocolMasterDNS:
		return NewMasterDNSService().RenameConfig(file.Name, newName)

	case DNSTunProtocolStormDNS:
		return NewStormDNSService().RenameConfig(file.Name, newName)

	case DNSTunProtocolTheFeed:
		return NewTheFeedService().RenameConfig(file.Name, newName)

	default:
		return fmt.Errorf("unsupported DNS tunnel protocol: %q", file.Protocol)
	}
}

// proxyLabel renders the "proxytype-authmethod" label used by the tunnel
// config listing.
func proxyLabel(proxyType ResolverProxyType, authMethod AuthMethod) string {
	return string(proxyType) + "-" + string(authMethod)
}
