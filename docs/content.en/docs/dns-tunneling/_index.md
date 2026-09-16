---
title: "DNS Tunneling"
weight: 3
bookFlatSection: true
bookCollapseSection: true
---

# DNS Tunneling

bgscan tests whether DNS resolvers can carry a tunnel connection. Six protocols are supported, each with its own configuration format stored as a separate TOML file under `assets/dns-tunneling/`.

Navigate to **Main Menu → DNS Tunneling** to manage tunnel configurations.

## Protocols

| Protocol | Description | Requires |
| --- | --- | --- |
| [**DNSTT**](./dnstt/) | DNS tunnel using the vaydns library with DNSTT-compatible framing | Domain, public key |
| [**VayDNS**](./vaydns/) | Native vaydns protocol with tunable QNAME, MTU, and record type | Domain, public key |
| [**Slipstream**](./slipstream/) | External `slipstream-client` binary based tunnel | Domain, binary on PATH |
| [**MasterDNS**](./masterdns/) | In-process tunnel with configurable encryption and MTU discovery | Domain, encryption key |
| [**StormDNS**](./stormdns/) | In-process tunnel with configurable encryption, MTU discovery, and query type selection | Domain, encryption key |
| [**TheFeed**](./thefeed/) | Encrypted feed tunnel over UDP, TCP, or DNS-over-TLS | Domain, passphrase |

DNSTT, VayDNS, MasterDNS, StormDNS, and TheFeed run the tunnel in-process using vendored libraries. Slipstream shells out to an external binary and communicates through a local SOCKS5 port.

DNSTT, VayDNS, and Slipstream support optional SOCKS5 or SSH proxy routing with password or key authentication. MasterDNS, StormDNS, and TheFeed configs carry no proxy settings.

The scan flow runs all six configurations. Pick any saved tunnel config and the matching stage is built automatically, with an optional resolver pre-scan ahead of it.

## How It Works

A DNS tunnel scan proceeds in two stages when `check_dns_resolver` is enabled:

1. **Resolver pre-scan** — each target IP is tested as a DNS resolver. Only resolvers that pass basic DNS queries (and the DPI check when enabled) proceed to the tunnel stage.
2. **Tunnel probe** — each surviving resolver is tested for its ability to carry a tunnel connection. The probe brings up the tunnel through the selected protocol and validates that traffic passes.

When `adaptive_resolver` is enabled, the resolver pre-scan automatically uses the same transport, port, and domain as the tunnel configuration, ensuring the resolver test matches the tunnel's actual path.

Reported latency measures the tunnel after it is up, excluding startup cost.

## Managing Configurations

The DNS Tunneling table at **Main Menu → DNS Tunneling** lists all saved tunnel configurations:

| Column | Description |
| --- | --- |
| Name | Configuration name |
| Protocol | DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, or TheFeed |
| Auth | Authentication method (none, password, key) |
| Created Time | File creation timestamp |

| Key | Action |
| --- | --- |
| `a` | Add a new configuration (opens protocol selector) |
| `r` | Rename the selected configuration |
| `x` | Delete the selected configuration |
| `Enter` | Edit or start a scan with the selected configuration |

## Configuration Storage

Tunnel configs are stored as TOML files:

```
assets/dns-tunneling/
├── dnstt/
│   └── <config-name>.toml
├── vaydns/
│   └── <config-name>.toml
├── slipstream/
│   └── <config-name>.toml
├── masterdns/
│   └── <config-name>.toml
├── stormdns/
│   └── <config-name>.toml
└── thefeed/
    └── <config-name>.toml
```

## Platform Defaults

Worker counts and timeouts are automatically adjusted based on the detected platform and resource tier:

| Setting | Desktop Low | Desktop Mid | Desktop High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 8 | 16 | 32 |
| DNS Tunnel Timeout | 10s | 10s | 10s |
| DNS Resolver Workers | 30 | 150 | 300 |

| Setting | Android Low | Android Mid | Android High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 3 | 6 | 12 |
| DNS Tunnel Timeout | 10s | 10s | 10s |
| DNS Resolver Workers | 15 | 60 | 100 |

| Setting | Server Low | Server Mid | Server High |
| --- | --- | --- | --- |
| DNS Tunnel Workers | 12 | 24 | 64 |
| DNS Tunnel Timeout | 10s | 10s | 8s |
| DNS Resolver Workers | 100 | 400 | 1000 |

## Related Topics

- [DNS Settings](../settings/dns.md) — Resolver and tunnel orchestration settings
- [Scan Types](../scanner/scan-types.md) — How DNS tunneling fits in the scan pipeline
- [Result Files](../scanner/result-files.md) — Tunnel result file format
