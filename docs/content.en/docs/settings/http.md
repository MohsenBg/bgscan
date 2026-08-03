---
title: "HTTP Settings"
weight: 7
---

# HTTP Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **HTTP Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/http_settings.toml`

The HTTP probe sends one request per target and records status code, negotiated protocol version, and whether TLS was used. HTTP/1.1 and HTTP/2 go over TCP, HTTP/3 over QUIC.

## Quick Reference

| Setting | Default | Description |
|---------|---------|-------------|
| `workers` | `50` | Concurrent requests (1-5000) |
| `host` | `"example.com"` | Request host, optionally with a path |
| `server_name` | `""` | SNI override; empty means derive from `host` |
| `port` | `443` | Target port (1-65535) |
| `protocol` | `"https"` | `http` or `https` |
| `version` | `"h1,h2"` | Protocol version to negotiate |
| `tls_validation` | `true` | Verify the server certificate |
| `min_tls_version` | `"tls1.1"` | Lowest TLS version allowed |
| `max_tls_version` | `"tls1.3"` | Highest TLS version allowed |
| `timeout` | `4000` | Request timeout in milliseconds (100-60000) |
| `accepted_status_codes` | `[]` | Status codes treated as success; empty accepts all |
| `prefix_output` | `"http_"` | Result filename prefix |

## Workers

```toml
workers = 50
```

Concurrent requests in flight. HTTP probes hold a connection and read a response, so they cost more per worker than TCP.

- `10-50` on limited resources
- `50-100` for typical use
- `100-200` for high-throughput scanning

## Host

```toml
host = "example.com"
```

Host used for the request. A path may be included (`example.com/path`) and is appended to the request URL. The `Host` header always carries only the domain part.

## Port

```toml
port = 443
```

Target port. Use 80 for plain HTTP, 443 for HTTPS, or any port the service listens on.

## Protocol

```toml
protocol = "https"
```

Either `"http"` or `"https"`. This decides whether the connection is wrapped in TLS.

## TLS Validation

```toml
tls_validation = true
```

When `true`, the certificate must be valid and trusted by the system. When `false`, expired and self-signed certificates are accepted. Turn it off when scanning IPs directly, where the certificate rarely matches the address.

## HTTP Version

```toml
version = "h1,h2"
```

| Value | Meaning |
|---|---|
| `h1` | HTTP/1.1 only |
| `h2` | HTTP/2 only |
| `h1,h2` | Negotiated over ALPN |
| `h3` | HTTP/3 over QUIC |

Longer aliases are also accepted: `http1`, `http1.1`, `http2`, `http3`, `http1,http2`, `http2,http1`. HTTP/3 uses a separate QUIC probe and always runs over TLS 1.3, so `min_tls_version` and `max_tls_version` are ignored.

## Timeout

```toml
timeout = 4000
```

Time budget for a single request, in milliseconds.

- `2000-5000` for reliable networks
- `5000-10000` for distant or unreliable networks

## TLS Version Range

```toml
min_tls_version = "tls1.1"
max_tls_version = "tls1.3"
```

Bounds passed to the TLS handshake. Accepted values are `tls1.0`, `tls1.1`, `tls1.2`, and `tls1.3`. The minimum must not exceed the maximum. Ignored when `version = "h3"`.

## Server Name Indication

```toml
server_name = ""
```

SNI sent during the TLS handshake. When empty, the domain from `host` is used. Set it to test domain fronting, or to probe an IP while presenting a specific hostname.

## Accepted Status Codes

```toml
accepted_status_codes = []
```

Allow-list of status codes counted as a successful result. An empty list, or a list covering every code, disables filtering and accepts anything the server returns. Responses outside the list are discarded and do not reach the next pipeline stage.

```toml
accepted_status_codes = [200, 204, 301, 302, 307, 308]
```

## Prefix Output

```toml
prefix_output = "http_"
```

Filename prefix for result files written by this probe. Files land in `result/http/`.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
