---
title: "DNS Settings"
weight: 8
---

# DNS settings

> **Note:** You do not need to edit the TOML file by hand. You can change these settings directly inside BGScan.
>
> Go to **Settings → DNS Settings** from the main menu and set the options with the Inspector inside the TUI.

Settings file: `settings/dns_settings.toml`

This file has two sections:

- **`resolver`:** Settings for scanning DNS resolvers.
- **`dns_tunneling`:** Settings for running the DNS tunnel scan.

The specific settings for each tunnel protocol, meaning DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed, are in separate files under `assets/dns-tunneling/`.

## Quick reference

| Setting | Default | Description |
|---|---|---|
| `resolver.workers` | Depends on the platform | The number of concurrent DNS queries (1 to 2500) |
| `resolver.protocol` | `"udp"` | The communication protocol: `udp`, `tcp`, or `dot` |
| `resolver.domain` | `"example.com"` | The domain asked of each resolver |
| `resolver.port` | `53` | The resolver port (1 to 65535) |
| `resolver.check_types` | `["TXT"]` | The record types checked in order |
| `resolver.edns_buffer_size` | `1232` | The EDNS0 UDP buffer size in bytes. A value of `0` disables it |
| `resolver.timeout` | Depends on the platform | The timeout for each query in milliseconds (100 to 30000) |
| `resolver.tries` | `1` | The number of attempts for each record type (1 to 10) |
| `resolver.random_subdomain` | `true` | Adds a random label to bypass the cache |
| `resolver.accepted_rcodes` | `["NOERROR","NXDOMAIN","SERVFAIL"]` | The response codes that show the resolver answers |
| `resolver.output_prefix` | `"dns_"` | The result file name prefix |
| `resolver.dpi.enabled` | `true` | Enables the DPI check before the main probe |
| `resolver.dpi.timeout` | `2000` | The DPI check timeout in milliseconds (100 to 10000) |
| `resolver.dpi.tries` | `1` | The number of attempts for the DPI check (1 to 10) |
| `dns_tunneling.workers` | Depends on the platform | The number of concurrent tunnel test workers (1 to 500) |
| `dns_tunneling.tries` | `1` | The number of attempts for each target (1 to 10) |
| `dns_tunneling.timeout` | Depends on the platform | The tunnel test timeout in milliseconds (100 to 60000) |
| `dns_tunneling.check_dns_resolver` | `true` | Runs a resolver scan before the tunnel test |
| `dns_tunneling.adaptive_resolver` | `true` | Matches the resolver settings to the tunnel config |
| `dns_tunneling.output_prefix` | `"dns_tun_"` | The result file name prefix |

## Resolver

In this scan, each target IP is checked as a resolver. The probe asks it about the set domain and keeps the target if the response code is in `accepted_rcodes`.

### Workers

```toml
[resolver]
workers = 150
```

This option sets the number of concurrent queries. The effective default depends on the platform and the available resources.

Because UDP queries are relatively cheap, you can set a higher worker count than for the HTTP scan. Of course, if you send a large number of requests to one upstream network, the scan traffic is easier to notice.

### Protocol

```toml
[resolver]
protocol = "udp"
```

The protocol used to send the queries. The three values `udp`, `tcp`, and `dot` are supported. The input value is not case sensitive and falls back to `udp` if it is unknown.

### Domain

```toml
[resolver]
domain = "example.com"
```

The domain asked of each resolver. Enter only the domain name, with no scheme, port, or path.

Choose a domain that resolves correctly on the internet. If the chosen domain cannot be resolved, healthy resolvers may also be marked invalid.

### Port

```toml
[resolver]
port = 53
```

The port the queries are sent to. DNS over UDP or TCP usually uses port `53`. If `protocol = "dot"`, the usual port is `853`.

### Record types to check

```toml
[resolver]
check_types = ["TXT"]
```

The record types the probe tries in order. The check stops as soon as the first answer with an acceptable Rcode arrives, and the successful record type is recorded in the result.

If some resolvers answer one record type but refuse another, you can put several record types in this list.

Supported types:

`A`, `AAAA`, `CNAME`, `NS`, `MX`, `TXT`, `SRV`, `NULL`, and `CAA`.

### EDNS buffer size

```toml
[resolver]
edns_buffer_size = 1232
```

Sets the UDP payload size announced in the EDNS0 OPT record. A value of `0` disables EDNS0.

### Timeout

```toml
[resolver]
timeout = 2000
```

The maximum wait for each query, in milliseconds.

### Number of attempts

```toml
[resolver]
tries = 1
```

Sets the number of attempts for each record type. Retries happen only for network errors.

As soon as a DNS answer arrives, even if its Rcode is not acceptable, the probe does not retry and moves on to the next record type.

### Random subdomain

```toml
[resolver]
random_subdomain = true
```

When this option is enabled, a random 10 character label is added to the domain for each probe.

This reduces the chance of using an answer stored in the resolver cache and makes the request need a real resolution instead. As a result, the response time gives more useful information about the real performance of the resolver.

### Acceptable Rcodes

```toml
[resolver]
accepted_rcodes = ["NOERROR", "NXDOMAIN", "SERVFAIL"]
```

This option defines which DNS answers are enough to consider the resolver alive.

| Value | Other names | Code |
|---|---|---:|
| `NOERROR` | `success` | 0 |
| `FORMERR` | `formaterror` | 1 |
| `SERVFAIL` | `serverfailure` | 2 |
| `NXDOMAIN` | `nameerror` | 3 |
| `NOTIMP` | `notimplemented` | 4 |
| `REFUSED` | — | 5 |

When `random_subdomain` is enabled, getting `NXDOMAIN` for a random subdomain is normal, because such a name usually does not exist. For this reason, `NXDOMAIN` is among the acceptable answers by default.

### Output prefix

```toml
[resolver]
output_prefix = "dns_"
```

This option sets the file name prefix for the resolver scan result files. Files are saved under `result/dns_resolver/`.

### DPI check

```toml
[resolver.dpi]
enabled = true
timeout = 2000
tries = 1
```

The DPI or Deep Packet Inspection check runs before the main queries. The probe asks for a random name under the `.invalid` domain, which cannot be registered for normal internet names.

Under the logic of this check, if the resolver returns a `NOERROR` answer for this name, there is a chance of answer tampering or DNS spoofing and the target is rejected. Getting any other Rcode means it passed this check.

The check result is recorded in each record as `passed` or `skipped`.

- `timeout`: The maximum wait in milliseconds, from 100 to 10000.
- `tries`: The number of attempts, from 1 to 10.

It is better to set the DPI check timeout lower than the main timeout so targets that do not answer are dropped faster.

## DNS tunnel

The `dns_tunneling` section controls the settings for running the DNS tunnel scan. The specific settings for each protocol are not in this section and you must use the TOML file for that protocol.

### Workers

```toml
[dns_tunneling]
workers = 16
```

Sets the number of concurrent tunnel test workers. Each worker runs its own tunnel probe and has a local port, so the tunnel test usually uses more resources than a simple resolver check.

The effective default depends on the platform and the available resources.

### Number of attempts

```toml
[dns_tunneling]
tries = 1
```

Sets the number of attempts for each target.

### Timeout

```toml
[dns_tunneling]
timeout = 10000
```

The maximum time assigned to bringing up the tunnel and validating it, in milliseconds.

Since bringing up a tunnel needs more steps than a simple DNS query, you usually need to give it a larger timeout.

### DNS resolver check

```toml
[dns_tunneling]
check_dns_resolver = true
```

If this option is enabled, a resolver pre-scan runs before the tunnel test starts. Only resolvers that pass this stage are checked for the tunnel test.

This way resources are not spent testing tunnels whose resolver does not even answer a simple DNS query.

### Adaptive resolver

```toml
[dns_tunneling]
adaptive_resolver = true
```

When this option is enabled, the resolver settings, including the protocol, port, and domain, are adjusted automatically to match the chosen tunnel config.

For example, if the DNSTT config uses `resolver_type = "tcp"` and port `853`, the resolver pre-scan also runs with that protocol and port. As a result, the pre-scan checks the path the tunnel is going to use.

If this option is disabled, the pre-scan uses the settings from the `[resolver]` section.

### Output prefix

```toml
[dns_tunneling]
output_prefix = "dns_tun_"
```

This option sets the file name prefix for the tunnel scan result files. Where the results are saved depends on the chosen protocol:

- `result/dnstt/`
- `result/vaydns/`
- `result/slipstream/`
- `result/masterdns/`
- `result/stormdns/`
- `result/thefeed/`

## Tunnel configs

Each DNS tunnel protocol, meaning DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed, keeps its own settings in separate TOML files under `assets/dns-tunneling/`.

The directory structure is:

```text
assets/dns-tunneling/
├── dnstt/
│   ├── my-dnstt-config.toml
│   └── ...
├── vaydns/
│   ├── my-vaydns-config.toml
│   └── ...
├── slipstream/
│   ├── my-slipstream-config.toml
│   └── ...
├── masterdns/
│   ├── my-masterdns-config.toml
│   └── ...
├── stormdns/
│   ├── my-stormdns-config.toml
│   └── ...
└── thefeed/
    ├── my-thefeed-config.toml
    └── ...
```

You can create and manage these configs through the TUI at **Main Menu → DNS Tunneling**. Each config has a name, a protocol type, and its own options.

To learn about the settings for each protocol, see the [DNS tunnel](../dns-tunneling/) page.

## Related files

- [`general_settings.toml`](./general.md): general scan settings and the scan execution mode
- [`icmp_settings.toml`](./icmp.md): ICMP scan settings
- [`tcp_settings.toml`](./tcp.md): TCP port scan settings
- [`http_settings.toml`](./http.md): HTTP, HTTPS, and HTTP/3 probe settings
- [`xray_settings.toml`](./xray.md): Xray scan settings and outbound validation
- [`writer_settings.toml`](./writer.md): result saving settings
- [DNS tunnel](../dns-tunneling/): the specific settings for the tunnel protocols
