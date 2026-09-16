---
title: "TheFeed"
weight: 6
---

# TheFeed Configuration

TheFeed runs in-process through a vendored client library. Queries and responses are encrypted with AES-256 keys derived from a shared passphrase. There are no proxy settings. The probe timeout comes from the global DNS tunnel settings (`dns_settings.toml`, `dns_tunneling.timeout`), not from this file.

## Connection Settings

| Field | Default | Description |
| --- | --- | --- |
| `Domain` | (required) | DNS subdomain that answers feed queries (for example `t.example.com`) |
| `Passphrase` | (required) | Same passphrase as the server; used to derive the AES-256 keys |
| `ResolverPort` | `53` | Resolver port (TheFeed servers often use 5300) |
| `ResolverType` | `'udp'` | Transport: `udp`, `tcp`, or `dot` |
| `QueryMode` | `'single'` | Query encoding: `single` (one base32 label) or `double` (multi-label hex) |

## Example Config

```toml
domain = 't.example.com'
passphrase = 'your-passphrase'
resolver_port = 53
resolver_type = 'udp'
query_mode = 'single'
```

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [DNS Settings](../../settings/dns.md) — Global tunnel timeout and resolver pre-scan
