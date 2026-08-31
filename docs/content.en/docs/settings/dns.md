---
title: "DNS Settings"
weight: 8
---

# DNS Settings

> **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **DNS Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/dns_settings.toml`

This file has two sections. The **resolver** section controls plain DNS resolver testing. The **dns_tunneling** section controls how DNS tunnel scans are orchestrated — tunnel protocol-specific settings (DNSTT, VayDNS, Slipstream) are stored in separate config files under `assets/dns-tunneling/`.

## Quick Reference

| Setting | Default | Description |
|---------|---------|-------------|
| `resolver.workers` | platform-dependent | Concurrent DNS queries (1-2500) |
| `resolver.protocol` | `"udp"` | Transport: `udp`, `tcp`, `dot` |
| `resolver.domain` | `"example.com"` | Domain queried through each resolver |
| `resolver.port` | `53` | Resolver port (1-65535) |
| `resolver.check_types` | `["TXT"]` | Record types tried in order |
| `resolver.edns_buffer_size` | `1232` | EDNS0 UDP buffer size in bytes; 0 disables |
| `resolver.timeout` | platform-dependent | Per-query timeout in ms (100-30000) |
| `resolver.tries` | `1` | Attempts per record type (1-10) |
| `resolver.random_subdomain` | `true` | Prepend a random label to bypass caches |
| `resolver.accepted_rcodes` | `["NOERROR","NXDOMAIN","SERVFAIL"]` | Response codes counted as alive |
| `resolver.output_prefix` | `"dns_"` | Result filename prefix |
| `resolver.dpi.enabled` | `true` | Run the hijack check before probing |
| `resolver.dpi.timeout` | `2000` | DPI check timeout in ms (100-10000) |
| `resolver.dpi.tries` | `1` | DPI check attempts (1-10) |
| `dns_tunneling.workers` | platform-dependent | Concurrent tunnel test workers (1-500) |
| `dns_tunneling.tries` | `1` | Retry attempts per target (1-10) |
| `dns_tunneling.timeout` | platform-dependent | Tunnel test timeout in ms (100-60000) |
| `dns_tunneling.check_dns_resolver` | `true` | Run resolver scan before tunnel test |
| `dns_tunneling.adaptive_resolver` | `true` | Override resolver settings to match tunnel config |
| `dns_tunneling.output_prefix` | `"dns_tun_"` | Result filename prefix |

## Resolver

Each target IP is treated as a resolver. The probe queries `domain` through it and keeps the target if the response code is in `accepted_rcodes`.

### Workers

```toml
[resolver]
workers = 150
```

Concurrent queries in flight. The effective default depends on platform and resource tier. UDP queries are cheap, so this scales further than the HTTP probe, but a high rate against a single upstream network will be noticed.

### Protocol

```toml
[resolver]
protocol = "udp"
```

Transport used for queries: `udp`, `tcp`, or `dot`. Parsing is case-insensitive; unknown values default to `udp`.

### Domain

```toml
[resolver]
domain = "example.com"
```

Domain queried through each resolver. It must be a bare domain with no scheme, port, or path. Pick a name that resolves reliably worldwide, otherwise honest resolvers will look broken.

### Port

```toml
[resolver]
port = 53
```

Port the resolver listens on. Use 853 with `protocol = "dot"`.

### Check Types

```toml
[resolver]
check_types = ["TXT"]
```

Record types tried in order. The probe stops at the first type that returns an accepted rcode and records that type in the result. Add more types when a resolver may answer for one and refuse another.

Supported record types: `A`, `AAAA`, `CNAME`, `NS`, `MX`, `TXT`, `SRV`, `NULL`, `CAA`.

### EDNS Buffer Size

```toml
[resolver]
edns_buffer_size = 1232
```

UDP payload size advertised in the OPT record. `0` disables EDNS0.

### Timeout

```toml
[resolver]
timeout = 2000
```

Per-query timeout in milliseconds.

### Tries

```toml
[resolver]
tries = 1
```

Attempts per record type. Retries only cover network errors. Once any DNS response arrives, even with a rejected rcode, the probe moves on to the next record type without retrying.

### Random Subdomain

```toml
[resolver]
random_subdomain = true
```

Prepends a random 10-character label to `domain` for each probe. This defeats resolver caches and forces a real recursive lookup, so latency reflects the resolver doing work rather than serving a cached answer.

### Accepted RCodes

```toml
[resolver]
accepted_rcodes = ["NOERROR", "NXDOMAIN", "SERVFAIL"]
```

Response codes counted as a live resolver.

| Value | Aliases | Code |
|---|---|---|
| `NOERROR` | `success` | 0 |
| `FORMERR` | `formaterror` | 1 |
| `SERVFAIL` | `serverfailure` | 2 |
| `NXDOMAIN` | `nameerror` | 3 |
| `NOTIMP` | `notimplemented` | 4 |
| `REFUSED` | | 5 |

With `random_subdomain` enabled, `NXDOMAIN` is a normal answer for a made-up label, which is why it is accepted by default.

### Output Prefix

```toml
[resolver]
output_prefix = "dns_"
```

Filename prefix for resolver results. Files land in `result/dns_resolver/`.

### DPI Check

```toml
[resolver.dpi]
enabled = true
timeout = 2000
tries = 1
```

The DPI (Deep Packet Inspection) check runs before the real queries. The probe asks the resolver for a random `.invalid` name, which cannot exist. A resolver that answers `NOERROR` is fabricating results, so the target is dropped. Any other rcode counts as honest. The outcome is stored per result as `passed` or `skipped`.

`timeout` is in milliseconds (range 100-10000). `tries` ranges from 1-10. Keep the DPI timeout well below the main `timeout` so dead targets are discarded quickly.

## DNS Tunneling

The `dns_tunneling` section controls how tunnel scans are orchestrated. It does not contain tunnel protocol settings — those are stored in separate per-config TOML files under `assets/dns-tunneling/`.

### Workers

```toml
[dns_tunneling]
workers = 16
```

Concurrent tunnel test workers. Each worker runs its own tunnel probe and holds a local port, so this is far more expensive than the resolver probe. The effective default depends on platform and resource tier.

### Tries

```toml
[dns_tunneling]
tries = 1
```

Retry attempts per target.

### Timeout

```toml
[dns_tunneling]
timeout = 10000
```

Time budget for bringing up the tunnel and validating it, in milliseconds. Tunnel tests need more time than plain DNS queries.

### Check DNS Resolver

```toml
[dns_tunneling]
check_dns_resolver = true
```

When `true`, a resolver pre-scan runs before the tunnel test. Only resolvers that pass the resolver scan are tested as tunnel candidates. This avoids wasting tunnel probes on resolvers that cannot even answer basic DNS queries.

### Adaptive Resolver

```toml
[dns_tunneling]
adaptive_resolver = true
```

When `true`, the resolver settings (transport, port, domain) are automatically overridden to match the selected tunnel configuration. For example, if a DNSTT config uses `resolver_type = "tcp"` on port 853, the resolver pre-scan will use the same transport and port. This ensures the resolver pre-scan tests the same path the tunnel will use.

When `false`, the resolver pre-scan uses the settings from the `[resolver]` section as-is.

### Output Prefix

```toml
[dns_tunneling]
output_prefix = "dns_tun_"
```

Filename prefix for tunnel result files. Files land in `result/dnstt/`, `result/vaydns/`, or `result/slipstream/` depending on the protocol.

## Tunnel Configurations

DNS tunnel protocols (DNSTT, VayDNS, Slipstream) are configured through separate TOML files stored under `assets/dns-tunneling/`:

```
assets/dns-tunneling/
├── dnstt/
│   ├── my-dnstt-config.toml
│   └── ...
├── vaydns/
│   ├── my-vaydns-config.toml
│   └── ...
└── slipstream/
    ├── my-slipstream-config.toml
    └── ...
```

These configs are created and managed through the TUI at **Main Menu → DNS Tunneling**. Each config has a name, a protocol type, and protocol-specific fields. See [DNS Tunneling](../dns-tunneling/) for detailed configuration of each protocol.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
- [DNS Tunneling](../dns-tunneling/) — Tunnel protocol configuration
