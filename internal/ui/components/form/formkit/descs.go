package formkit

// Shared field descriptions reused across the tunnel configuration forms.
// Protocol-specific fields (e.g. MTU tuning) keep their local descriptions.
const (
	DescConfigName    = "The name of the configuration file."
	DescDomain        = "The target domain name for the tunnel."
	DescPubKey        = "The server public key used for encryption."
	DescPassphrase    = "The encryption passphrase shared with the server."
	DescResolverType  = "The DNS resolver transport: UDP, TCP, or DOT."
	DescResolverPort  = "The DNS resolver port (typically 53)."
	DescFingerprint   = "The TLS fingerprint for the resolver."
	DescRPS           = "Requests per second (0 = unlimited)."
	DescEncryptionKey = "The shared encryption key."
	DescEncMethod     = "The data encryption method."
	DescDNSQueryType  = "The DNS query shape: TXT, NS, CNAME or ROTATE."
	DescQueryMode     = "The DNS query encoding: single (base32) or double (multi-label hex)."
	DescProxyType     = "The proxy type: SOCKS or SSH."
	DescProxyPort     = "The proxy port number."
	DescAuthMethod    = "The authentication method: none, password, or key."
	DescUsername      = "The username for authentication."
	DescPassword      = "The password for password authentication."
	DescPrivateKey    = "The SSH private key used for authentication."

	// MTU tuning fields (masterdns / stormdns).
	DescMTUTestTimeout      = "MTU test timeout in seconds."
	DescMTUTestRetries      = "MTU test retries."
	DescSessionInitRetryMax = "Maximum session init retry time in seconds."
	DescSessionInitRacing   = "Session init racing count."
	DescMinUploadMTU        = "Minimum upload MTU."
	DescMaxUploadMTU        = "Maximum upload MTU."
	DescMinDownloadMTU      = "Minimum download MTU."
	DescMaxDownloadMTU      = "Maximum download MTU."
	DescMTUParallelism      = "MTU test parallelism."
	DescRxTxWorkers         = "RX/TX worker count."
)
