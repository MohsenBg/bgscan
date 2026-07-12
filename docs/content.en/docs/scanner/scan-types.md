---
title: "Scan Types"
weight: 2
---

# Scan Types

bgscan supports five primary scan types, each accessible from the main menu via keyboard shortcuts:

| Key | Scan Type | Description | Primary Use Case | Configuration File |
|-----|-----------|-------------|------------------|-------------------|
| `i` | **ICMP Scan** | Sends ICMP echo requests (ping) to check host availability | Host discovery, network mapping | [`icmp_settings.toml`](../settings/icmp.md) |
| `t` | **TCP Scan** | Attempts TCP connections to specified ports | Port scanning, service discovery | [`tcp_settings.toml`](../settings/tcp.md) |
| `h` | **HTTP Scan** | Sends HTTP/HTTPS requests to web servers | Web service testing, API endpoint discovery | [`http_settings.toml`](../settings/http.md) |
| `d` | **DNS Scan** | Queries DNS servers for resolution and tunneling capabilities | DNS resolver testing, tunneling detection | [`dns_settings.toml`](../settings/dns.md) |
| `x` | **Xray Scan** | Tests outbound connectivity and bandwidth to remote servers | Network egress testing, bandwidth measurement | [`xray_settings.toml`](../settings/xray.md) |

> 💡 **Note**: The DNS Scan option (`d`) actually encompasses three sub-scan types when enabled in configuration:
>
> - **DNS Resolve**: Standard DNS queries (A/AAAA/TXT records)
> - **DNSTT**: DNS Tunnel Tool tests for tunneling capability
> - **SlipStream**: Alternative DNS tunneling detection
>
> These sub-scans are controlled individually in the DNS settings file.

{{< img "/bgscan-scan-type.webp" "bgscan scan type" >}}

## Detailed Scan Type Information

#### ICMP Scan

- **Protocol**: ICMP Echo Request/Reply (ping)
- **What it does**: Sends echo requests to target IPs and measures round-trip time for replies
- **Output**: Lists responsive hosts with response times
- **Key settings**: Timeout, retry attempts, worker concurrency
- **Best for**: Initial host discovery before running more intensive scans

#### TCP Scan

- **Protocol**: TCP SYN/connect scan
- **What it does**: Attempts to establish TCP connections to specified ports on target IPs
- **Output**: Lists hosts with open ports and connection times
- **Key settings**: Target port, timeout, retry attempts, worker concurrency
- **Best for**: Service discovery, port scanning, firewall testing

#### HTTP Scan

- **Protocol**: HTTP/HTTPS (supports HTTP/1.1, HTTP/2, HTTP/3 via ALPN)
- **What it does**: Sends HTTP requests to target hosts and ports, evaluates responses
- **Output**: Lists responsive web servers with status codes, response times, and optional headers/content
- **Key settings**: Target host/port, protocol (HTTP/HTTPS), TLS validation, HTTP version, accepted status codes
- **Best for**: Web service testing, API endpoint discovery, load balancer checking

#### DNS Scan

- **Protocol**: DNS queries (UDP, TCP, or DNS-over-TLS)
- **What it does**: Sends DNS queries to target resolvers and evaluates their responses
- **Output**: Lists working DNS resolvers with response times and capabilities
- **Key settings**: Resolver workers, protocol, domain, query types, timeout, retries, accepted response codes, anti-hijacking checks
- **Best for**: Finding open/resolving DNS servers, testing DNS security configurations

#### Xray Scan

- **Protocol**: Custom Xray protocol for connectivity and bandwidth testing
- **What it does**: Tests outbound network connectivity and measures upload/download speeds to configured servers
- **Output**: Lists reachable servers with connection times and measured bandwidth
- **Key settings**: Timeout, worker count, test type (connect/download/upload/both), data sizes for transfer tests
- **Best for**: Network egress testing, bandwidth measurement, proxy/chokepoint detection
