---
title: "DNSTT"
weight: 1
---

# DNSTT Configuration

DNSTT uses the vaydns library with DNSTT-compatible framing (DNS-over-TXT with legacy framing). The tunnel stack is: DNS packet channel → KCP → Noise encryption → smux multiplexing.

## Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone delegated to your DNSTT server |
| `PubKey` | (required) | 64 hex chars | Server public key |
| `ResolverType` | `"udp"` | `udp`, `tcp`, `dot` | Transport to resolver |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `Fingerprint` | `"Chrome"` | see below | uTLS ClientHello fingerprint |
| `RPS` | `0` | 0-500 | Rate limit (queries/sec); 0 = unlimited |

## Proxy and Authentication

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

## TLS Fingerprints

The `Fingerprint` field selects a uTLS ClientHello profile to mimic during the TLS handshake. Labels are matched case-insensitively.

| Category | Labels |
| --- | --- |
| Chrome | `Chrome`, `Chrome_58`, `Chrome_62`, `Chrome_70`, `Chrome_72`, `Chrome_83`, `Chrome_87`, `Chrome_96`, `Chrome_100`, `Chrome_102`, `Chrome_120` |
| Firefox | `Firefox`, `Firefox_55`, `Firefox_56`, `Firefox_63`, `Firefox_65`, `Firefox_99`, `Firefox_102`, `Firefox_105`, `Firefox_120` |
| iOS | `iOS`, `iOS_11_1`, `iOS_12_1`, `iOS_13`, `iOS_14` |
| Other | `random` |

## Example Config

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

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [VayDNS](../vaydns/) — Native vaydns protocol with the same proxy options
