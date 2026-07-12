---
title: "Result Files"
weight: 5
---

# Result Files

After a scan completes, bgscan writes the discovered IPs and their measurements to result files. These files are organized by created time and can be browsed, viewed, renamed, or deleted from the TUI.

---

## Viewing Result Files

Navigate to **Main Menu → Result Files** to open the result file browser.

The table lists every result file across all scan types, sorted newest-first, and shows:

| Column | Description |
|---|---|
| File Name | The result file name (without `.csv`) |
| Created Time | File creation or last modification timestamp |
| Type | The scan engine that produced the file (`icmp`, `tcp`, `http`, etc.) |
| Size | File size on disk |

Two keyboard actions are available:

| Key | Action |
|---|---|
| `r` | Rename the selected file |
| `x` | Delete the selected file permanently |

![bgscan result file](/bgscan-result.webp)

#### Opening a result file

Press Enter on any file to open the IP viewer, which displays the IPs found in that file along with their measured latency, download, and upload values. Xray results show full download and upload measurements; all other scan types show a short view with latency only.

---

## Storage Location

Result files are stored under the `result/` directory next to the bgscan binary, in a subdirectory per scan type:

```
<bgscan-root>/
└── result/
    ├── icmp/
    ├── tcp/
    ├── http/
    ├── xray/
    ├── resolve/
    ├── dnstt/
    └── slipstream/
```

Each directory holds `.csv` files produced by its corresponding scan engine. Files are created automatically when a scan runs — you do not need to create these directories manually.

---

## File Naming

Result files are named using the scan engine's configured prefix followed by a timestamp:

```
<prefix><YYYYMMDD_HHMMSS>.csv
```

Each scan type has a default prefix set in its configuration:

| Scan Type | Default Prefix | Example Filename |
|---|---|---|
| ICMP | `icmp_` | `icmp_20240711_143022.csv` |
| TCP | `tcp_` | `tcp_20240711_143022.csv` |
| HTTP | `http_` | `http_20240711_143022.csv` |
| Xray | `xray_` | `xray_20240711_143022.csv` |
| DNS Resolver | `dns_resolver_` | `dns_resolver_20240711_143022.csv` |
| DNSTT | `dns_dnstt_` | `dns_dnstt_20240711_143022.csv` |
| SlipStream | `dns_slipstream_` | `dns_slipstream_20240711_143022.csv` |

You can change the prefix for any scan type in its settings file under `settings/`. See [Configuration](../configuration.md) for details.

---

## File Format

Result files are plain CSV with no header row. Each row represents one responsive IP:

```
<ip>,<latency>,<download>,<upload>
```

| Field | Description |
|---|---|
| `ip` | The IPv4 address that responded |
| `latency` | Round-trip or connection latency (e.g. `123ms`) |
| `download` | Download measurement — `0s` when not measured |
| `upload` | Upload measurement — `0s` when not measured |

#### Example

```csv
1.2.3.4,45ms,0s,0s
5.6.7.8,120ms,320ms,80ms
```

For scan types that do not measure speed (ICMP, TCP, HTTP, DNS), the download and upload columns are present but set to `0s`.

---

## Result Ordering

IPs within a result file are sorted by a quality score — higher is better. The score is calculated from all three measurements:

- **60%** download speed
- **20%** upload speed  
- **20%** latency

This means the best IPs (fastest download, lowest latency) appear at the top of the file. When two IPs have the same score, they are sorted alphabetically by IP address as a tie-breaker.

---

## Related Topics

- [Scanner Overview](../scanner.md)
- [Scan Types](./scan-types.md)
- [IP Lists](./ip-files.md)
- [Scan Pipeline](./scan-pipeline.md)
