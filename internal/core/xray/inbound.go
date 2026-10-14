package xray

// getInbound builds the localhost SOCKS inbound used for probing.
func getInbound(port uint16) Inbound {
	return Inbound{
		Tag:      "socks-inbound",
		Listen:   "127.0.0.1",
		Port:     port,
		Protocol: ProtocolSocks,
		Settings: SocksSettings{
			Auth: SocksAuthNone,
			IP:   "127.0.0.1",
		},
	}
}
