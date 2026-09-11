package xray

// getInbound builds the localhost SOCKS inbound probes dial through.
// Auth is off (local use only); sniffing is on for http/tls detection.
func getInbound(port uint16) Inbound {
	return Inbound{
		Port:     port,
		Listen:   "127.0.0.1",
		Tag:      "socks-inbound",
		Protocol: "socks",
		Settings: SocksSettings{
			Auth: "noauth",
			UDP:  false,
			IP:   "127.0.0.1",
		},
		Sniffing: SniffingSetting{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}
}
