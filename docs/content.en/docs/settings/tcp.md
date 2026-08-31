---
title: "TCP Settings"
weight: 6
---

# TCP Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **TCP Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/tcp_settings.toml`

The TCP probe opens a connection to one port on each target and records the handshake latency. It does not speak any application protocol.

## Quick Reference

| Setting | Default | Range | Description |
| --------- | --------- | ------- | ------------- |
| `port` | `443` | 1-65535 | Target TCP port |
| `timeout` | `2000` | 100-30000 | Connection timeout in milliseconds |
| `workers` | `400` | 1-5000 | Concurrent connection attempts |
| `tries` | `1` | 1-10 | Connection attempts per target |
| `output_prefix` | `"tcp_"` | | Result filename prefix |

## Port

```toml
port = 80
```

The TCP port probed on every target. Common choices are 80, 443, and 22.

## Timeout

```toml
timeout = 3000
```

How long to wait for the handshake, in milliseconds. Targets that do not complete within the window count as unreachable.

- `1000-3000` for local networks
- `5000-10000` for distant or unreliable networks

## Workers

```toml
workers = 200
```

Concurrent connection attempts. Each worker holds one open socket, so high values need a matching file-descriptor limit.

- `50-100` on limited resources
- `200-500` for typical use
- `500-1000` for high-throughput scanning

## Retries

```toml
tries = 1
```

Attempts per target before giving up. Only timeouts are retried. A refused connection or other immediate error fails the target right away, which keeps throughput up on dead ranges. The minimum is 1, so there is no way to disable the first attempt.

## Output Prefix

```toml
output_prefix = "tcp_"
```

Filename prefix for result files written by this probe. Files land in `result/tcp/`.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`icmp_settings.toml`](./icmp.md) — ICMP scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
