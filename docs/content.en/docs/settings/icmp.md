---
title: "ICMP Settings"
weight: 5
---

# ICMP Settings

> 💡 **Tip:** You can change these settings directly in the bgscan application instead of editing the TOML file manually.
>
> Navigate to **Settings** → **ICMP Settings** in the main menu to configure these options interactively using the TUI inspector.

Configuration file: `settings/icmp_settings.toml`

The ICMP probe sends echo requests and records the round-trip time. It opens one shared IPv4 socket and, when the system allows it, one shared IPv6 socket, then demultiplexes replies to the waiting workers.

## Quick Reference

| Setting | Default | Range | Description |
| --------- | --------- | ------- | ------------- |
| `timeout` | `2000` | 100-30000 | Time to wait for an echo reply, in milliseconds |
| `tries` | `1` | 1-10 | Echo requests per target |
| `workers` | `200` | 1-5000 | Concurrent probes in flight |
| `output_prefix` | `"icmp_"` | | Result filename prefix |

## Timeout

```toml
timeout = 2000
```

How long a single echo request waits for its reply, in milliseconds. Applies per attempt, so the worst case for a target is `timeout * tries`.

## Retries

```toml
tries = 1
```

Number of echo requests sent before a target is declared unreachable. The attempt count that produced the reply is stored in the result.

## Workers

```toml
workers = 200
```

Number of targets probed concurrently. Workers share the same sockets, so raising this mostly affects packet rate rather than file descriptors. Very high values can trigger rate limiting on intermediate routers.

## Output Prefix

```toml
output_prefix = "icmp_"
```

Filename prefix for result files written by this probe. Files land in `result/icmp/`.

## Socket Mode

There is no setting for this, but it shows up in the results. The probe first tries to open a raw ICMP socket. If the process lacks the privileges, it falls back to an unprivileged UDP (datagram) ICMP socket. The mode actually used is recorded per result as `raw` or `udp`.

IPv6 support is best effort. If the IPv6 socket cannot be opened, IPv4 scanning still works and IPv6 targets fail with an error.

## Related Files

- [`general_settings.toml`](./general.md) — Global scan control and pipeline mode
- [`tcp_settings.toml`](./tcp.md) — TCP port scan configuration
- [`http_settings.toml`](./http.md) — HTTP/HTTPS/HTTP3 probe configuration
- [`dns_settings.toml`](./dns.md) — DNS scan configuration
- [`xray_settings.toml`](./xray.md) — Xray outbound validation
- [`writer_settings.toml`](./writer.md) — Result output configuration
