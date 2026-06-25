<div align="right">
  
  [**English**](./dns_settings.md) &nbsp;|&nbsp; [**فارسی**](./dns_settings.fa.md)
  
</div>

# DNS Scanner Configuration

Configuration file: `settings/dns_settings.toml`

This file controls all DNS-based scanning and tunnelling modules: the **Bulk Resolver Scanner**, the **DNSTT Probe**, and the **Slipstream Probe**.

---

## Table of Contents

- [Resolver — Bulk DNS Scanner](#resolver--bulk-dns-scanner)
  - [Core](#core)
  - [Query](#query)
  - [Timing & Retry](#timing--retry)
  - [Filtering](#filtering)
  - [DPI / Anti-Hijacking](#dpi--anti-hijacking)
  - [Output](#output)
- [DNSTT — DNS Tunnel Probe](#dnstt--dns-tunnel-probe)
- [Slipstream — DNS Tunnel Probe](#slipstream--dns-tunnel-probe)

---

## `[resolver]` — Bulk DNS Scanner

Tests public DNS resolvers to find those that are reachable and behave correctly. Results feed into subsequent probes (DNSTT, Slipstream).

---

### Core

#### `workers`

```toml
workers = 500
```

Number of concurrent goroutines performing resolver scans.

| Recommended Range | Note                              |
| ----------------- | --------------------------------- |
| `50` – `500`      | Scale up with available bandwidth |

---

#### `protocol`

```toml
protocol = "udp"
```

Transport protocol used for DNS queries.

| Value   | Description                |
| ------- | -------------------------- |
| `"udp"` | Standard UDP DNS (default) |
| `"tcp"` | DNS over TCP               |
| `"dot"` | DNS-over-TLS               |

> **Note:** `"doh"` (DNS-over-HTTPS) is **not supported**.

---

#### `domain`

```toml
domain = ""
```

The primary domain name used as the query target. Must be set before running.

---

#### `port`

```toml
port = 53
```

DNS port used when scanning resolvers. Standard DNS port is `53`.

---

### Query

#### `check_types`

```toml
check_types = ["txt"]
```

DNS record types to query during scanning.

| Common Values | Description                                                |
| ------------- | ---------------------------------------------------------- |
| `"txt"`       | TXT records — strongly recommended for DNSTT compatibility |
| `"a"`         | IPv4 address records                                       |
| `"aaaa"`      | IPv6 address records                                       |
| `"ns"`        | Nameserver records                                         |
| `"mx"`        | Mail exchange records                                      |

---

#### `ends_buffer_size`

```toml
ends_buffer_size = 0
```

EDNS buffer size (in bytes) sent in OPT records. Set to `0` to use the resolver's default.

---

### Timing & Retry

#### `timeout`

```toml
timeout = 2000
```

Time in **milliseconds** to wait for a DNS response before considering the resolver unresponsive.

---

#### `tries`

```toml
tries = 2
```

Number of network-level retries for the main probe.

> **Important:** Retries only occur on network errors (timeouts, connection failures). If **any** DNS response is received — regardless of Rcode — retrying stops immediately.

---

#### `random_subdomain`

```toml
random_subdomain = true
```

When enabled, prepends a random subdomain to the query (e.g., `x1y2z3.example.com`).

**Why enable this?**

- Bypasses resolver-side caching
- Forces a fresh recursive lookup upstream
- Ensures results reflect live resolver behavior

---

### Filtering

#### `accepted_rcodes`

```toml
accepted_rcodes = ["noerror", "nxdomain"]
```

Defines which DNS response codes are considered a sign that the resolver is **alive**.

| Value                            | Code | Meaning                     |
| -------------------------------- | ---- | --------------------------- |
| `"noerror"` / `"success"`        | `0`  | Query answered successfully |
| `"formerr"` / `"formaterror"`    | `1`  | Malformed request           |
| `"servfail"` / `"serverfailure"` | `2`  | Server-side failure         |
| `"nxdomain"` / `"nameerror"`     | `3`  | Domain does not exist       |
| `"notimp"` / `"notimplemented"`  | `4`  | Query type not supported    |
| `"refused"`                      | `5`  | Query refused by resolver   |

> **Recommended:** `["noerror", "nxdomain"]`

---

### DPI / Anti-Hijacking

Detects resolvers that tamper with DNS responses (e.g., ISP-level hijacking). The check sends a query for a non-existent `.invalid` domain — if the resolver returns `NOERROR` (Rcode 0), it is flagged as a hijacking resolver and discarded.

#### `check_dpi`

```toml
check_dpi = false
```

Enable or disable the anti-hijacking pre-check.

---

#### `dpi_timeout`

```toml
dpi_timeout = 500
```

Timeout in **milliseconds** for the DPI check. Intentionally shorter than `timeout` to quickly skip bad resolvers.

---

#### `dpi_tries`

```toml
dpi_tries = 2
```

Max network retries for the DPI check. Only retries on timeout or network-level errors.

---

### Output

#### `prefix_output`

```toml
prefix_output = "dns_"
```

Prefix added to the generated output file name (e.g., `dns_results.txt`).

---

## `[dnstt]` — DNS Tunnel Probe

Tests whether resolvers that survived the bulk scan can carry a full DNS tunnel. Requires a [DNSTT](https://www.bamsoftware.com/software/dnstt/) server running on the authoritative nameserver.

| Setting         | Default    | Description                                           |
| --------------- | ---------- | ----------------------------------------------------- |
| `enabled`       | `false`    | Enable or disable this probe phase                    |
| `workers`       | `20`       | Concurrent handshake workers (recommended: `5`–`50`)  |
| `domain`        | `""`       | Authoritative DNS zone delegated to your DNSTT server |
| `public_key`    | `""`       | Base64-encoded Ed25519 public key of the DNSTT server |
| `timeout`       | `10000`    | Milliseconds allowed for a complete handshake attempt |
| `prefix_output` | `"dnstt_"` | Prefix for the output file name                       |

**Example:**

```toml
[dnstt]
enabled    = true
workers    = 20
domain     = "tunnel.example.com"
public_key = "base64encodedpublickey=="
timeout    = 10000
```

---

## `[slip_stream]` — Slipstream DNS Probe

An alternative tunnelling technique that exploits DNS behavior. Operates independently of DNSTT and can run alongside it.

| Setting         | Default         | Description                                               |
| --------------- | --------------- | --------------------------------------------------------- |
| `enabled`       | `true`          | Enable or disable this probe phase                        |
| `workers`       | `20`            | Concurrent probe workers (recommended: `5`–`50`)          |
| `domain`        | `""`            | Authoritative DNS zone used by the Slipstream server      |
| `cert_path`     | `""`            | Path to the TLS certificate used by the Slipstream server |
| `timeout`       | `8000`          | Milliseconds allowed for a complete probe attempt         |
| `prefix_output` | `"slipstream_"` | Prefix for the output file name                           |

**Example:**

```toml
[slip_stream]
enabled     = true
workers     = 20
domain      = "slip.example.com"
cert_path   = "/etc/certs/slipstream.crt"
timeout     = 8000
```

---

## Full Example

```toml
[resolver]
workers           = 500
protocol          = "udp"
domain            = "example.com"
port              = 53
check_types       = ["txt"]
ends_buffer_size  = 0
timeout           = 2000
tries             = 2
random_subdomain  = true
accepted_rcodes   = ["noerror", "nxdomain"]
check_dpi         = true
dpi_timeout       = 500
dpi_tries         = 2
prefix_output     = "dns_"

[dnstt]
enabled       = true
workers       = 20
domain        = "tunnel.example.com"
public_key    = "base64encodedpublickey=="
timeout       = 10000
prefix_output = "dnstt_"

[slip_stream]
enabled       = true
workers       = 20
domain        = "slip.example.com"
cert_path     = "/etc/certs/slipstream.crt"
timeout       = 8000
prefix_output = "slipstream_"
```
