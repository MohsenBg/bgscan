---
title: "Slipstream"
weight: 3
---

# Slipstream Configuration

Slipstream runs as an external process (`slipstream-client` binary), not an in-process Go tunnel. It communicates through a local SOCKS5 port allocated per probe.

## Connection Settings

| Field | Default | Range | Description |
| --- | --- | --- | --- |
| `Domain` | (required) | valid domain | Zone served by your Slipstream server |
| `ResolverPort` | `53` | > 0 | Resolver port |
| `CertPath` | `""` | path | Optional TLS certificate (passed as `--cert`) |

## Proxy and Authentication

Same fields as DNSTT (see [Proxy and Authentication](../dnstt/#proxy-and-authentication)).

## Binary Location

bgscan searches for the `slipstream-client` binary in the following locations:

1. `<bgscan-root>/assets/slipstream-client/slipstream-client`
2. `<bgscan-root>/assets/slipstream/slipstream-client/slipstream-client`
3. `<bgscan-root>/slipstream-client/slipstream-client`
4. `<bgscan-root>/slipstream-client`
5. System `PATH`

## Example Config

```toml
Domain = "example.com"
ResolverPort = 53
CertPath = ""
ProxyType = "socks"
ProxyPort = 1080
AuthMethod = "none"
```

## Related Topics

- [DNS Tunneling](../) — Protocol overview and scan flow
- [DNSTT](../dnstt/) — Proxy and authentication fields
