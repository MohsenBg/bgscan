---
title: "VayDNS"
weight: 2
---

# VayDNS Configuration

VayDNS is the native vaydns protocol. It uses the same tunnel stack as DNSTT (DNS → KCP → Noise → smux) but with native framing and additional tuning parameters for QNAME structure, MTU, and record type.

## Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone delegated to your VayDNS server |
| `PubKey` | (required) | 64 hex chars | Server public key |
| `ResolverType` | `"udp"` | `udp`, `tcp`, `dot` | Transport to resolver |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `Fingerprint` | `"Chrome"` | see DNSTT fingerprints | uTLS ClientHello fingerprint |
| `RecordType` | `"TXT"` | see below | DNS record type for tunnel queries |
| `RPS` | `0` | 0-500 | Rate limit (queries/sec); 0 = unlimited |

## Advanced Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `ClientIDSize` | `2` | 1-8 | Client ID byte size |
| `MaxQnameLen` | `101` | 0-253 | Max QNAME length in DNS packets (0 = auto) |
| `MaxNumLabels` | `0` | 0-4 | Max number of labels in QNAME (0 = auto) |
| `MTU` | `0` | 0-1452 | Max transmission unit (0 = auto) |

## DNS Record Types

Supported `RecordType` values: `A`, `AAAA`, `CNAME`, `NS`, `MX`, `TXT`, `SRV`, `NULL`, `CAA`.

The record type determines which DNS record type is used for tunnel queries. `TXT` is the default and most widely supported.

## Proxy and Authentication

Same fields as DNSTT (see [Proxy and Authentication](../dnstt/#proxy-and-authentication)).

## Example Config

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

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [DNSTT](../dnstt/) — DNSTT-compatible framing and TLS fingerprints
