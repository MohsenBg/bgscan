---
title: "Scan Types"
weight: 2
---

# Scan types

The scan type menu has six options, each chosen with a key:

| Key | Scan type | What it does | Settings file |
|-----|-----------|--------------|--------------------|
| `i` | ICMP scan | Sends an echo request and measures the round trip time | [`icmp_settings.toml`](../settings/icmp.md) |
| `t` | TCP scan | Connects to a port and measures the handshake latency | [`tcp_settings.toml`](../settings/tcp.md) |
| `h` | HTTP scan | Sends an HTTP request and records the response status, version, and TLS | [`http_settings.toml`](../settings/http.md) |
| `x` | Xray scan | Passes traffic through an outbound and checks the connection | [`xray_settings.toml`](../settings/xray.md) |
| `r` | DNS resolver | Tests each IP as a resolver and checks for DPI hijacking when needed | [`dns_settings.toml`](../settings/dns.md) |
| `d` | DNS tunnel | Checks which resolvers can carry a DNS tunnel | [`dns_settings.toml`](../settings/dns.md) |

**DNS resolver** is a standalone scan. It checks each target IP as a resolver and, when the matching option is enabled, also detects DPI hijacking.

**DNS tunnel** first opens the config browser so you can choose one of the saved configs: DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, or TheFeed. If the `check_dns_resolver` option is enabled in the DNS tunnel settings, a resolver pre-scan also runs before the tunnel starts.

**Xray scan** first asks you to choose the outbound template. If `pre_scan_type` is `icmp`, `tcp`, or `http`, that scan runs before Xray and filters the targets. With `none`, the Xray scan starts without a pre-scan.

{{< img "/bgscan-scan-type.webp" "bgscan scan type" >}}

## ICMP

This probe sends an echo request and waits for a reply. It uses one shared IPv4 socket across the workers and, if the system allows, a shared IPv6 socket as well. The reply to each request goes back to the same worker that sent it.

It records the latency, the number of attempts, and the socket type. The socket can be raw or, when the needed access is missing, it can use UDP.

ICMP is a low cost way to narrow a large range of IPs before running heavier scans.

## TCP

The TCP scan is a full connect scan, not a SYN scan. That means it carries the connection to the configured port through the end of the handshake and records the latency of the first successful attempt.

If the connection times out, it tries again up to the `tries` limit. If the connection is refused, it fails right away so checking inactive IPs finishes faster.

The result has the latency, the port, and the attempt count.

## HTTP

For each target it sends a request with HTTP/1.1, HTTP/2, or HTTP/3. The choice between `h1` and `h2` happens through ALPN, but `h3` has its own probe that uses QUIC.

Responses whose status code is not in `accepted_status_codes` are not accepted. When this list is empty, every status code is acceptable.

The result has the latency, the status code, the chosen HTTP version, and whether TLS was used.

## Xray

For each target, an Xray instance runs with the chosen outbound template. Latency is then measured through the local proxy.

Depending on the `connectivity_test_type` value, a download test, an upload test, or both also run. If the speed is lower than the configured minimum, that target is rejected.

The result has the latency and the download and upload speed. If only the connection is checked, both speed values stay at zero.

## DNS resolver

In this scan, each target IP is checked as a DNS resolver. If `check_dpi` is enabled, the probe first asks for a random name under the `.invalid` domain. If the resolver returns a successful answer for this invalid name, it is probably tampering with DNS answers, so that resolver is dropped from the rest of the scan.

After this step, the probe asks the remaining resolvers about the configured record types. If one of the requests gets an acceptable response code, the resolver is accepted.

The result has the latency, the record type that answered, the attempt count, the response code, and whether the DPI check ran.

## DNS tunnel

This scan checks which resolvers can carry the traffic of a DNS tunnel. First you must choose one of the saved configs from the config browser.

All six protocols below can be scanned:

- **DNSTT:** A DNS tunnel using the VayDNS library and framing compatible with the old DNSTT version
- **VayDNS:** The native VayDNS protocol with settings for QNAME, MTU, and record type
- **Slipstream:** The internal tunnel client `libslipstream-client` with settings for DNS resolution and congestion control
- **TheFeed:** An encrypted tunnel over UDP, TCP, or DNS-over-TLS with single query or double query encoding
- **MasterDNS:** An internal tunnel with configurable encryption and MTU discovery
- **StormDNS:** An internal tunnel with configurable encryption, MTU discovery, and a choice of query type

Slipstream, MasterDNS, and StormDNS run their own tunnel client for each target and listen on a local SOCKS5 port. For this reason, scanning them is much slower than the resolver check stage.

DNSTT and VayDNS validate the connection through the tunnel itself. In all these protocols, latency is measured after the tunnel is up, not the time it took to bring the tunnel up.

If `check_dns_resolver` is enabled, a resolver pre-scan runs before the tunnel and only active targets reach the tunnel stage. The `adaptive_resolver` option also adjusts the resolver settings to match the transport, port, and domain in the tunnel config.

For details on the settings for each protocol, see the [DNS tunnel](../dns-tunneling/) page.

## Related topics

- [Scan source](./scan-source.md)
- [Scan stages](./scan-pipeline.md)
- [Result files](./result-files.md)
- [DNS tunnel](../dns-tunneling/)
