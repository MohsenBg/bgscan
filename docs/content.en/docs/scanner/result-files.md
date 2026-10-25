---
title: "Result Files"
weight: 5
---

# Result files

When the scan finishes, BGScan saves the successful addresses and the information it measured in a result file. Files go into separate directories based on the scan type. From inside the TUI you can also see the files, rename them, or delete them.

---

## Viewing result files

Go to **Result Files** from the main menu (`r`) to see the list of result files.

The table shows the files registered in the result registry. Files are sorted from newest to oldest and display this information:

| Column | Description |
|---|---|
| File Name | The result file name without the `.csv` extension |
| Created Time | The time the file was last changed |
| Type | The result type, such as `ICMP`, `TCP`, `HTTP`, `DNSResolver`, and `DNSTT` |
| Size | The file size on disk |

The main keys in this section:

| Key | Action |
|---|---|
| `Enter` | Open the selected file in the IP Viewer |
| `r` | Rename the file |
| `x` | Delete the file permanently |

{{< img "/bgscan-result.webp" "bgscan result file" >}}

### Table keys

The result file browser uses the same shared table as the rest of the program:

| Key | Action |
|---|---|
| `↑` / `↓` or `k` / `j` | Move up or down one row |
| `u` / `d` | Move by half a page |
| `PgUp` / `PgDn` | Move up or down one page |
| `g` / `G` | Go to the first or last row |
| `s` | Change the sort column |
| `o` | Reverse the sort order |
| `/` | Filter the rows |
| `b` or `Esc` | Go back to the previous screen |

Sorting follows the data type of each column. For example, IPs, times, numbers, and text each sort in their natural way. Time and bandwidth are formatted for display, but sorting uses their real values.

### Opening a result file

Press `Enter` on the file you want to open it in the IP Viewer. The displayed columns depend on the result type. For example, ICMP, TCP, HTTP, DNS, and tunnel results have their own details, and Xray results show download and upload speed in addition to latency.

Time and bandwidth values are also colored based on their quality.

> **Current limitation:** MasterDNS and StormDNS result files are saved to disk, but their schemas are not registered when the program starts. As a result, these files do not appear in the Result Files browser and cannot be opened from there. You can still open their CSV files with other tools that support CSV.

---

## Where files are stored

Result files are stored in the `result/` directory next to the BGScan binary. Each scan type has its own directory:

```text
<bgscan-root>/
└── result/
    ├── icmp/
    ├── tcp/
    ├── http/
    ├── xray/
    ├── dns_resolver/
    ├── dnstt/
    ├── vaydns/
    ├── slipstream/
    ├── masterdns/
    ├── stormdns/
    └── thefeed/
```

Each directory holds the `.csv` files for that scan type. These directories are created automatically when a scan runs, so you do not need to create them by hand.

---

## Result file names

Each file name is built from the prefix set for the scan type and the time it was created:

```text
<prefix><YYYYMMDD_HHMMSS>.csv
```

Each scan type has its own default prefix:

| Scan type | Default prefix | Example file name |
|---|---|---|
| ICMP | `icmp_` | `icmp_20240711_143022.csv` |
| TCP | `tcp_` | `tcp_20240711_143022.csv` |
| HTTP | `http_` | `http_20240711_143022.csv` |
| Xray | `xray_` | `xray_20240711_143022.csv` |
| DNS resolver | `dns_` | `dns_20240711_143022.csv` |
| DNSTT | `dns_tun_` | `dns_tun_20240711_143022.csv` |
| VayDNS | `dns_tun_` | `dns_tun_20240711_143022.csv` |
| Slipstream | `dns_tun_` | `dns_tun_20240711_143022.csv` |
| MasterDNS | `dns_tun_` | `dns_tun_20240711_143022.csv` |
| StormDNS | `dns_tun_` | `dns_tun_20240711_143022.csv` |
| TheFeed | `dns_tun_` | `dns_tun_20240711_143022.csv` |

You can change the prefix for each scan type in the settings file for that type. For more information, see the [Settings overview](../settings/overview.md).

---

## Structure of result files

Result files are stored in CSV format and have no header row. Each row holds the details of one target that the scan accepted. The structure of each row depends on the schema for that scan type.

Every scan type saves at least the target IP and the latency. The other columns depend on what each probe measures.

| Scan type | CSV columns |
|---|---|
| ICMP | `ip, latency, tries, mode` |
| TCP | `ip, latency, port, tries` |
| HTTP | `ip, latency, status, version, tls` |
| Xray | `ip, latency, download, upload` |
| DNS resolver | `ip, latency, record_type, tries, rcode, dpi_check` |
| DNSTT | `ip, latency, transport, port, auth, proxy` |
| VayDNS | `ip, latency, transport, port, auth, proxy` |
| Slipstream | `ip, latency, port, auth, proxy` |
| MasterDNS | `ip, latency, port, enc` |
| StormDNS | `ip, latency, port, query, enc` |
| TheFeed | `ip, latency, domain, mode` |

What the columns mean:

- `ip`: The IPv4 or IPv6 address of the target that responded.
- `latency`: The round trip or connection latency, for example `123ms`.
- `download` / `upload`: The download and upload speed in an Xray scan. When the matching test did not run, the value is `0 bps`.
- `status` / `version` / `tls`: The HTTP status code, the agreed HTTP version, and whether TLS was used.
- `tries`: The number of attempts until a successful result.
- `mode`: The ICMP socket type (`raw` or `udp`) or the TheFeed query mode.
- `rcode` / `dpi_check`: The DNS response code and the result of the DPI check for detecting tampered answers.
- `transport`: The DNS transport type in DNSTT or VayDNS, such as `udp`, `tcp`, or `dot`.
- `port`: The resolver port in DNSTT, VayDNS, MasterDNS, and StormDNS, or the local SOCKS listener port in Slipstream.
- `auth` / `proxy`: The authentication method and proxy type set for DNSTT, VayDNS, or Slipstream.
- `enc`: The encryption method number in MasterDNS or StormDNS.
- `query`: The query type in StormDNS.
- `domain`: The domain used in TheFeed.

#### HTTP file example

```csv
1.2.3.4,45ms,200,HTTP/2.0,true
2606:4700::6810:85e5,120ms,403,HTTP/1.1,true
```

#### Xray file example

```csv
1.2.3.4,45ms,0 bps,0 bps
5.6.7.8,120ms,10.00 Mbps,5.00 Mbps
```

In scan types that do not measure speed, the download and upload columns do not exist.

---

## Order of results

Targets in a result file are sorted by the quality score for each schema. A higher score means a better result.

Each scan type has its own `Score()` method:

- **ICMP, TCP, HTTP, DNS, DNSTT, VayDNS, Slipstream, MasterDNS, StormDNS, and TheFeed:** The score is based on the inverse latency (`1000 / latency_ms`). Targets with lower latency therefore appear higher.
- **Xray:** The score combines latency (10 percent), download speed (60 percent), and upload speed (30 percent). If one of the speed tests is disabled, its share of the score becomes zero. In an Xray scan that only checks the connection, the result order is therefore based entirely on latency.

When saving new results, BGScan merges them with the existing file by score using Merge Sort. If an IP is a duplicate, the newer record replaces the earlier one. The order of entries with the same score is also defined in a fixed way so repeated merges produce the same result.

---

## Related topics

- [Scan types](./scan-types.md)
- [Scan source](./scan-source.md)
- [IP lists](./ip-files.md)
- [Scan stages](./scan-pipeline.md)
