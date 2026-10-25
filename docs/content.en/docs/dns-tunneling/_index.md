---
title: "DNS Tunneling"
weight: 3
bookFlatSection: true
bookCollapseSection: true
---

# DNS tunnel

BGScan checks which DNS resolvers can carry the traffic of a tunnel. Six protocols are supported for now, each with its own settings. These settings are stored in separate TOML files under the `assets/dns-tunneling/` directory.

To create and manage tunnel configs, go to **Main Menu → DNS Tunneling**.

## Protocols

| Protocol | Description | Requirements |
| --- | --- | --- |
| [**DNSTT**](./dnstt/) | A DNS tunnel using the VayDNS library and DNSTT compatible framing | Domain, public key |
| [**VayDNS**](./vaydns/) | The native VayDNS protocol with settings for QNAME, MTU, and record type | Domain, public key |
| [**Slipstream**](./slipstream/) | An internal tunnel client that loads the `libslipstream-client` library | Domain, Slipstream library |
| [**MasterDNS**](./masterdns/) | An internal tunnel with configurable encryption and MTU discovery | Domain, encryption key |
| [**StormDNS**](./stormdns/) | An internal tunnel with encryption, MTU discovery, and a choice of query type | Domain, encryption key |
| [**TheFeed**](./thefeed/) | An encrypted tunnel over UDP, TCP, or DNS-over-TLS | Domain, passphrase |

All six protocols run inside BGScan itself. Slipstream uses its own client library, which loads through `purego`, and the rest use internal Go libraries.

Slipstream, MasterDNS, and StormDNS start a local SOCKS5 listener for each probe.

DNSTT, VayDNS, and Slipstream support optional routing of traffic through a SOCKS5 or SSH proxy. These proxies can authenticate with a Password or a Key. The MasterDNS, StormDNS, and TheFeed configs have no proxy settings.

For each scan you must choose a saved config. BGScan builds the matching scan stage based on the chosen protocol. If the related option is enabled, a resolver pre-scan also runs before the tunnel test.

## How it works

If the `check_dns_resolver` option is enabled, the DNS tunnel scan runs in two stages:

1. **Resolver pre-scan:** Each target IP is checked as a DNS resolver. Only resolvers that answer basic DNS queries and, when the DPI check is enabled, pass that check move on to the next stage.
2. **Tunnel test:** Each remaining resolver is checked to find out whether it can carry the traffic of a tunnel connection. The probe brings up the tunnel with the chosen protocol and checks that traffic passes through it.

If `adaptive_resolver` is enabled, the resolver pre-scan settings are adjusted automatically to match the transport, port, and domain of the tunnel config. This way the pre-scan checks the same path the tunnel is going to use.

The latency reported in the result is measured after the tunnel is up, so the time spent bringing the tunnel up is not included in this value.

## Managing configs

The table in **Main Menu → DNS Tunneling** shows all saved configs.

| Column | Description |
| --- | --- |
| Name | Config name |
| Protocol | The protocol type: DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, or TheFeed |
| Auth | A summary of the proxy and authentication details for DNSTT, VayDNS, and Slipstream, or the encryption and Passphrase label for the other protocols |
| Created Time | The time the file was created |

The available keys:

| Key | Action |
| --- | --- |
| `a` | Create a new config and open the protocol picker |
| `r` | Rename the selected config |
| `x` | Delete the selected config |
| `Enter` | Edit the selected config |

## Where configs are stored

Tunnel configs are stored as TOML files in a directory for each protocol:

```text
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

## Platform default settings

BGScan adjusts the number of workers and the timeout automatically based on the platform and the system resource level.

**Desktop**

| Setting | Low | Mid | High |
| --- | ---: | ---: | ---: |
| DNS tunnel workers | 8 | 16 | 32 |
| DNS tunnel timeout | 10s | 10s | 10s |
| DNS resolver workers | 30 | 80 | 100 |

**Android**

| Setting | Low | Mid | High |
| --- | ---: | ---: | ---: |
| DNS tunnel workers | 3 | 6 | 12 |
| DNS tunnel timeout | 10s | 10s | 10s |
| DNS resolver workers | 15 | 40 | 80 |

**Server**

| Setting | Low | Mid | High |
| --- | ---: | ---: | ---: |
| DNS tunnel workers | 12 | 24 | 64 |
| DNS tunnel timeout | 10s | 10s | 8s |
| DNS resolver workers | 80 | 100 | 200 |

## Related topics

- [DNS settings](../settings/dns.md): the resolver settings and how they match the tunnel
- [Scan types](../scanner/scan-types.md): where the DNS tunnel scan fits in the scan stages
- [Result files](../scanner/result-files.md): the structure of the tunnel scan result files
