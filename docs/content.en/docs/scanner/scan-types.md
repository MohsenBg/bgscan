---
title: "Scan Types"
weight: 2
---

# Scan Types

The scan type menu offers five entries, each bound to a key:

| Key | Scan Type | What it does | Configuration File |
|-----|-----------|--------------|--------------------|
| `i` | ICMP Scan | Sends echo requests and measures round-trip time | [`icmp_settings.toml`](../settings/icmp.md) |
| `t` | TCP Scan | Opens a connection to one port and measures handshake latency | [`tcp_settings.toml`](../settings/tcp.md) |
| `h` | HTTP Scan | Sends one HTTP request and records status, version, and TLS | [`http_settings.toml`](../settings/http.md) |
| `d` | DNS Scan | Queries each target as a resolver, optionally tests tunnels | [`dns_settings.toml`](../settings/dns.md) |
| `x` | Xray Scan | Routes traffic through an Xray outbound and measures it | [`xray_settings.toml`](../settings/xray.md) |

Two of these expand into more than one pipeline stage.

**DNS Scan** always adds the resolver stage. It appends a DNSTT stage when `dnstt.enabled` is true, and a SlipStream stage when `slip_stream.enabled` is true. Each stage writes its own result file.

**Xray Scan** first asks which outbound template to use. If `pre_scan_type` is `icmp`, `tcp`, or `http`, that stage runs ahead of the Xray stage and filters the targets. With `none`, Xray runs directly.

{{< img "/bgscan-scan-type.webp" "bgscan scan type" >}}

## ICMP

Echo request and reply. The probe shares one IPv4 socket and, when the system allows, one IPv6 socket across all workers, matching replies back to the worker that sent them.

It records latency, the number of attempts, and whether the socket was raw or unprivileged UDP. Cheapest way to narrow a large range before a heavier scan.

## TCP

A full connect scan, not a SYN scan. It completes the handshake to the configured port and records the latency of the first successful attempt. Timeouts are retried up to `tries`, while a refused connection fails immediately so dead ranges drain quickly.

Records latency, port, and attempt count.

## HTTP

One request per target over HTTP/1.1, HTTP/2, or HTTP/3. Version selection happens over ALPN for `h1,h2`; `h3` uses a separate QUIC probe. Responses outside `accepted_status_codes` are discarded, and an empty list accepts everything.

Records latency, status code, negotiated version, and whether TLS was used.

## DNS

Each target IP is treated as a resolver. With `check_dpi` enabled, the probe first asks for a random `.invalid` name. A resolver that claims success for a name that cannot exist is fabricating answers and gets dropped. Surviving targets are then queried for the configured record types until one returns an accepted response code.

Records latency, the record type that worked, attempt count, response code, and whether the DPI check ran.

The DNSTT and SlipStream stages go further and check whether a resolver can actually carry a tunnel. Both need an external client binary and allocate a local SOCKS5 port per probe, so they are much slower than the resolver stage. Their latency measures the tunnel once it is up, not the time to bring it up.

## Xray

Starts a temporary Xray process per target using the selected outbound template, then measures latency through the local proxy. Depending on `connectivity_test_type`, it also runs a download test, an upload test, or both, and fails the target when the measured speed falls below the configured minimum.

Records latency plus download and upload throughput. Both speeds are zero when only connectivity is tested.

## Related Topics

- [Scan Source](./scan-source.md)
- [Scan Pipeline](./scan-pipeline.md)
- [Result Files](./result-files.md)
