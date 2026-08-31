---
title: "DNS Tunneling"
weight: 7
---

# DNS Tunneling

bgscan tests whether DNS resolvers can carry a tunnel connection. Three protocols are supported, each with its own configuration format stored as a separate TOML file under `assets/dns-tunneling/`.

Navigate to **Main Menu → DNS Tunneling** to manage tunnel configurations.

## Protocols

| Protocol | Description | Requires |
| --- | --- | --- |
| **DNSTT** | DNS tunnel using the vaydns library with DNSTT-compatible framing | Domain, public key |
| **VayDNS** | Native vaydns protocol with tunable QNAME, MTU, and record type | Domain, public key |
| **Slipstream** | External `slipstream-client` binary based tunnel | Domain, binary on PATH |

DNSTT and VayDNS run the tunnel in-process using the vaydns library. Slipstream shells out to an external binary and communicates through a local SOCKS5 port.

All three protocols support optional SOCKS5 or SSH proxy routing with password or key authentication.

## How It Works

A DNS tunnel scan proceeds in two stages when `check_dns_resolver` is enabled:

1. **Resolver pre-scan** — each target IP is tested as a DNS resolver. Only resolvers that pass basic DNS queries (and the DPI check when enabled) proceed to the tunnel stage.
2. **Tunnel probe** — each surviving resolver is tested for its ability to carry a tunnel connection. The probe establishes a full tunnel stack (DNS packet channel → KCP transport → Noise encryption → smux multiplexing) and validates the connection.

When `adaptive_resolver` is enabled, the resolver pre-scan automatically uses the same transport, port, and domain as the tunnel configuration, ensuring the resolver test matches the tunnel's actual path.

Reported latency measures the tunnel after it is up, excluding startup cost.

## Managing Configurations

The DNS Tunneling table at **Main Menu → DNS Tunneling** lists all saved tunnel configurations:

| Column | Description |
| --- | --- |
| Name | Configuration name |
| Protocol | DNSTT, VayDNS, or Slipstream |
| Auth | Authentication method (none, password, key) |
| Created Time | File creation timestamp |

| Key | Action |
| --- | --- |
| `a` | Add a new configuration (opens protocol selector) |
| `r` | Rename the selected configuration |
| `x` | Delete the selected configuration |
| `Enter` | Edit or start a scan with the selected configuration |

## Configuration Storage

Tunnel configs are stored as TOML files:

```
assets/dns-tunneling/
├── dnstt/
│   └── <config-name>.toml
├── vaydns/
│   └── <config-name>.toml
└── slipstream/
    └── <config-name>.toml
```

---

## DNSTT Configuration

DNSTT uses the vaydns library with DNSTT-compatible framing (DNS-over-TXT with legacy framing). The tunnel stack is: DNS packet channel → KCP → Noise encryption → smux multiplexing.

### Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone delegated to your DNSTT server |
| `PubKey` | (required) | 64 hex chars | Server public key |
| `ResolverType` | `"udp"` | `udp`, `tcp`, `dot` | Transport to resolver |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `Fingerprint` | `"Chrome"` | see below | uTLS ClientHello fingerprint |
| `RPS` | `0` | 0-500 | Rate limit (queries/sec); 0 = unlimited |

### Proxy and Authentication

| Field | Default | Description |
| --- | --- | --- |
| `ProxyType` | `"socks"` | Proxy type: `socks` or `ssh` |
| `ProxyPort` | `1080` | Proxy port |
| `AuthMethod` | `"none"` | Auth: `none`, `password`, or `key` |
| `Username` | `""` | Required for password/key auth |
| `Password` | `""` | Required for password auth |
| `PrivateKey` | `""` | PEM-encoded SSH private key (required for key auth) |
| `KnownHostsFile` | `""` | Path to SSH known_hosts file (optional, key auth only) |

**Proxy rules:**

- SSH proxy requires authentication (password or key).
- SOCKS proxy does not allow key authentication.

### TLS Fingerprints

The `Fingerprint` field selects a uTLS ClientHello profile to mimic during the TLS handshake. Labels are matched case-insensitively.

| Category | Labels |
| --- | --- |
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

### Example Config

```toml
Domain = "example.com"
PubKey = "0000000000000000000000000000000000000000000000000000000000000000"
ResolverType = "udp"
ResolverPort = 53
Fingerprint = "Chrome"
RPS = 0
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

---

## VayDNS Configuration

VayDNS is the native vaydns protocol. It uses the same tunnel stack as DNSTT (DNS → KCP → Noise → smux) but with native framing and additional tuning parameters for QNAME structure, MTU, and record type.

### Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone delegated to your VayDNS server |
| `PubKey` | (required) | 64 hex chars | Server public key |
| `ResolverType` | `"udp"` | `udp`, `tcp`, `dot` | Transport to resolver |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `Fingerprint` | `"Chrome"` | see above | uTLS ClientHello fingerprint |
| `RecordType` | `"TXT"` | see below | DNS record type for tunnel queries |
| `RPS` | `0` | 0-500 | Rate limit (queries/sec); 0 = unlimited |

### Advanced Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `ClientIDSize` | `2` | 1-8 | Client ID byte size |
| `MaxQnameLen` | `101` | 0-253 | Max QNAME length in DNS packets (0 = auto) |
| `MaxNumLabels` | `0` | 0-4 | Max number of labels in QNAME (0 = auto) |
| `MTU` | `0` | 0-1452 | Max transmission unit (0 = auto) |

### DNS Record Types

Supported `RecordType` values: `A`, `AAAA`, `CNAME`, `NS`, `MX`, `TXT`, `SRV`, `NULL`, `CAA`.

The record type determines which DNS record type is used for tunnel queries. `TXT` is the default and most widely supported.

### Proxy and Authentication

Same fields as DNSTT (see [Proxy and Authentication](#proxy-and-authentication) above).

### Example Config

```toml
Domain = "example.com"
PubKey = "0000000000000000000000000000000000000000000000000000000000000000"
ResolverType = "udp"
ResolverPort = 53
Fingerprint = "Chrome"
RecordType = "TXT"
ClientIDSize = 2
MaxQnameLen = 101
MaxNumLabels = 0
MTU = 0
RPS = 0
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

---

## Slipstream Configuration

Slipstream runs as an external process (`slipstream-client` binary), not an in-process Go tunnel. It communicates through a local SOCKS5 port allocated per probe.

### Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone served by your Slipstream server |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `CertPath` | `""` | path | Optional TLS certificate (passed as `--cert`) |

### Proxy and Authentication

Same fields as DNSTT (see [Proxy and Authentication](#proxy-and-authentication) above).

### Binary Location

bgscan searches for the `slipstream-client` binary in the following locations:

1. `<bgscan-root>/assets/slipstream-client/slipstream-client`
2. `<bgscan-root>/assets/slipstream/slipstream-client/slipstream-client`
3. `<bgscan-root>/slipstream-client/slipstream-client`
4. `<bgscan-root>/slipstream-client`
5. System `PATH`

### Example Config

```toml
Domain = "example.com"
ResolverPort = 53
CertPath = ""
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

---

## Platform Defaults

Worker counts and timeouts are automatically adjusted based on the detected platform and resource tier:

| Setting | Desktop Low | Desktop Mid | Desktop High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 8 | 16 | 32 |
| DNS Tunnel Timeout | 10s | 10s | 10s |
| DNS Resolver Workers | 30 | 150 | 300 |

| Setting | Android Low | Android Mid | Android High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 3 | 6 | 12 |
| DNS Tunnel Timeout | 10s | 10s | 10s |
| DNS Resolver Workers | 15 | 60 | 100 |

| Setting | Server Low | Server Mid | Server High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 12 | 24 | 64 |
| DNS Tunnel Timeout | 10s | 10s | 8s |
| DNS Resolver Workers | 100 | 400 | 1000 |

---

## Related Topics

- [DNS Settings](../settings/dns.md) — Resolver and tunnel orchestration settings
- [Scan Types](../scanner/scan-types.md) — How DNS tunneling fits in the scan pipeline
- [Result Files](../scanner/result-files.md) — Tunnel result file format
