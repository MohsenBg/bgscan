---
title: "DNS Settings"
weight: 8
---

# DNS Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **DNS Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/dns_settings.toml`

Three independent modules live in this file. The resolver section tests plain DNS resolvers. The DNSTT and SlipStream sections test whether a resolver can carry a tunnel, and both need an external client binary.

## Quick Reference

| Setting | Default | Description |
|---------|---------|-------------|
| `resolver.workers` | `100` | Concurrent DNS queries (1-2500) |
| `resolver.protocol` | `"udp"` | Transport: `udp`, `tcp`, `dot`, `doh` |
| `resolver.domain` | `"google.com"` | Domain queried through each resolver |
| `resolver.port` | `53` | Resolver port |
| `resolver.check_types` | `["A"]` | Record types tried in order |
| `resolver.ends_buffer_size` | `1234` | Advertised EDNS0 UDP buffer size |
| `resolver.timeout` | `2000` | Per-query timeout in milliseconds (100-30000) |
| `resolver.tries` | `1` | Attempts per record type (1-10) |
| `resolver.random_subdomain` | `true` | Prepend a random label to bypass caches |
| `resolver.accepted_rcodes` | `["noerror", "nxdomain"]` | Response codes counted as alive |
| `resolver.check_dpi` | `true` | Run the hijack check before probing |
| `resolver.dpi_timeout` | `500` | DPI check timeout in milliseconds (100-10000) |
| `resolver.dpi_tries` | `2` | DPI check attempts (1-10) |
| `resolver.prefix_output` | `"dns_resolver_"` | Result filename prefix |
| `dnstt.enabled` | `false` | Enable the DNSTT stage |
| `dnstt.workers` | `20` | Concurrent DNSTT handshakes (1-500) |
| `dnstt.domain` | `"ns.example.com"` | Zone delegated to your DNSTT server |
| `dnstt.public_key` | 64 zeros | Server public key, 64 hex characters |
| `dnstt.timeout` | `8000` | Handshake timeout in milliseconds |
| `dnstt.prefix_output` | `"dns_dnstt_"` | Result filename prefix |
| `slip_stream.enabled` | `false` | Enable the SlipStream stage |
| `slip_stream.workers` | `20` | Concurrent SlipStream probes (1-500) |
| `slip_stream.domain` | `"ns.example.com"` | Zone used by the SlipStream server |
| `slip_stream.cert_path` | `""` | Optional TLS certificate for the client |
| `slip_stream.timeout` | `8000` | Probe timeout in milliseconds |
| `slip_stream.prefix_output` | `"dns_slipstream_"` | Result filename prefix |

## Resolver

Each target IP is treated as a resolver. The probe queries `domain` through it and keeps the target if the response code is in `accepted_rcodes`.

### Workers

```toml
workers = 100
```

Concurrent queries in flight. UDP queries are cheap, so this scales further than the HTTP probe, but a high rate against a single upstream network will be noticed.

### Protocol

```toml
protocol = "udp"
```

Transport used for queries: `udp`, `tcp`, `dot`, or `doh`. DoH is accepted by the config but resolves to DoT at runtime, because DoH needs a domain-based endpoint while the scanner targets resolvers by IP.

### Domain

```toml
domain = "google.com"
```

Domain queried through each resolver. It must be a bare domain with no scheme, port, or path. Pick a name that resolves reliably worldwide, otherwise honest resolvers will look broken.

### Port

```toml
port = 53
```

Port the resolver listens on. Use 853 with `protocol = "dot"`.

### Check Types

```toml
check_types = ["A"]
```

Record types tried in order. The probe stops at the first type that returns an accepted rcode and records that type in the result. Add more types when a resolver may answer for one and refuse another.

### EDNS Buffer Size

```toml
ends_buffer_size = 1234
```

UDP payload size advertised in the OPT record. `0` disables EDNS0. Note the key is spelled `ends_buffer_size` in the file.

### Timeout

```toml
timeout = 2000
```

Per-query timeout in milliseconds.

### Tries

```toml
tries = 1
```

Attempts per record type. Retries only cover network errors. Once any DNS response arrives, even with a rejected rcode, the probe moves on to the next record type without retrying.

### Random Subdomain

```toml
random_subdomain = true
```

Prepends a random 10-character label to `domain` for each probe. This defeats resolver caches and forces a real recursive lookup, so latency reflects the resolver doing work rather than serving a cached answer.

### Accepted RCodes

```toml
accepted_rcodes = ["noerror", "nxdomain"]
```

Response codes counted as a live resolver.

| Value | Aliases | Code |
|---|---|---|
| `noerror` | `success` | 0 |
| `formerr` | `formaterror` | 1 |
| `servfail` | `serverfailure` | 2 |
| `nxdomain` | `nameerror` | 3 |
| `notimp` | `notimplemented` | 4 |
| `refused` | | 5 |

With `random_subdomain` enabled, `nxdomain` is a normal answer for a made-up label, which is why it is accepted by default.

### Check DPI

```toml
check_dpi = true
```

Runs before the real queries. The probe asks the resolver for a random `.invalid` name, which cannot exist. A resolver that answers `NOERROR` is fabricating results, so the target is dropped. Any other rcode counts as honest. The outcome is stored per result as `passed` or `skipped`.

### DPI Timeout and Tries

```toml
dpi_timeout = 500
dpi_tries = 2
```

Timeout and attempt count for the hijack check. Keep the timeout well below the main `timeout` so dead targets are discarded quickly.

### Prefix Output

```toml
prefix_output = "dns_resolver_"
```

Filename prefix for resolver results. Files land in `result/dns_resolver/`.

## DNSTT

Tests whether a resolver can carry a DNSTT tunnel. The stage shells out to `dnstt-client`, which must be present in `assets/` or on `PATH`, and probes the tunnel through a locally allocated SOCKS5 port. When the binary is missing, startup logs a warning and the scan type is disabled.

Reported latency measures the tunnel after it is up, excluding startup cost.

### Enabled

```toml
enabled = false
```

### Workers

```toml
workers = 20
```

Each worker runs its own client process and holds a local port, so this is far more expensive than the resolver probe.

### Domain

```toml
domain = "ns.example.com"
```

The zone delegated to your DNSTT server. Passed to the client as the tunnel domain.

### Public Key

```toml
public_key = "0000000000000000000000000000000000000000000000000000000000000000"
```

Server public key, passed through as `-pubkey`. It must be exactly 64 hexadecimal characters. The default is a placeholder and will not connect to anything.

### Timeout

```toml
timeout = 8000
```

Time budget for bringing up the tunnel and validating it, in milliseconds.

### Prefix Output

```toml
prefix_output = "dns_dnstt_"
```

Files land in `result/dnstt/`.

## SlipStream

An alternative DNS tunneling technique with the same shape as DNSTT: an external `slipstream-client` binary, a local SOCKS5 port per probe, and latency measured after the tunnel is established.

### Enabled

```toml
enabled = false
```

### Workers

```toml
workers = 20
```

### Domain

```toml
domain = "ns.example.com"
```

The zone served by your SlipStream server.

### Certificate Path

```toml
cert_path = ""
```

Optional TLS certificate file, passed to the client as `--cert`. Leave empty when the server does not require one.

### Timeout

```toml
timeout = 8000
```

### Prefix Output

```toml
prefix_output = "dns_slipstream_"
```

Files land in `result/slipstream/`.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
